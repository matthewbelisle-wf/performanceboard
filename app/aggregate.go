package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func makeAggregateDtoList(metrics []AggregateMetric) []JsonResponse {
	aggDtoList := []JsonResponse{}
	for _, metric := range metrics {
		aggDTO := make(JsonResponse)
		aggDTO["start"] = metric.StartTime
		aggDTO["min"] = metric.Min
		aggDTO["max"] = metric.Max
		aggDTO["mean"] = metric.Mean
		aggDTO["count"] = metric.Count
		aggDtoList = append(aggDtoList, aggDTO)
	}
	return aggDtoList
}

func readAggregates(context appengine.Context,
	boardKey *datastore.Key, namespace string, binType string,
	newestTime time.Time, duration time.Duration,
	limit int, cursor string) ([]AggregateMetric, error) {
	// newestTime - same as 'end'
	// duration - 0 if it should go forever, TODO::USE CURSOR!
	// binTypes ["day", "hour", "minute", "second"]
	q := datastore.NewQuery(AggregateMetricKind).
		Filter("Namespace =", namespace).
		Filter("BinType =", binType).
		Filter("StartTime <=", newestTime).
		Order("-StartTime").
		Ancestor(boardKey)

	if duration > 0 {
		oldestTime := newestTime.Add(-duration)
		q = q.Filter("StartTime >=", oldestTime)
	}

	//TODO use a limit and return a cursor
	var aggregates []AggregateMetric
	if _, err := q.GetAll(context, &aggregates); err != nil {
		context.Errorf("Error reading aggregates: %v", err)
		return nil, err
	}
	return aggregates, nil
}

// HTTP handlers

func getAggregates(writer http.ResponseWriter, request *http.Request) {
	context := appengine.NewContext(request)
	encodedKey := mux.Vars(request)["board"]
	boardKey, err := datastore.DecodeKey(encodedKey)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	namespace := mux.Vars(request)["namespace"]
	binType := mux.Vars(request)["bin_type"] // day, hour, minute, second

	end := time.Now()

	// parse optional end time
	if endParam := request.FormValue("end"); len(endParam) > 0 {
		if end, err = time.Parse(time.RFC3339, endParam); err != nil {
			context.Errorf("Error parsing end:%s:%v", endParam, err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
	}

	duration := time.Duration(0)

	// parse optional start time
	if startParam := request.FormValue("start"); len(startParam) > 0 {
		if start, err := time.Parse(time.RFC3339, startParam); err != nil {
			context.Errorf("Error parsing start:%s:%v", startParam, err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		} else {
			duration = end.Sub(start)
		}
	}

	// read aggregates
	aggregates, err := readAggregates(context, boardKey, namespace, binType, time.Now(), duration, 0, "")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	aggDtoList := makeAggregateDtoList(aggregates)
	b, err := json.Marshal(aggDtoList)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Write(b)
}
