package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

func makeAggregateDtoList(metrics []*AggregateMetric) []JsonResponse {
	aggDtoList := []JsonResponse{}
	for _, metric := range metrics {
		aggDTO := make(JsonResponse)
		aggDTO["start"] = (*metric).StartTime
		aggDTO["min"] = (*metric).Min
		aggDTO["max"] = (*metric).Max
		aggDTO["mean"] = (*metric).Mean
		aggDTO["count"] = (*metric).Count
		aggDtoList = append(aggDtoList, aggDTO)
	}
	return aggDtoList
}

func readAggregates(context appengine.Context,
	boardKey *datastore.Key, namespace string, binType string,
	newestTime time.Time, duration time.Duration,
	limit int, cursor string) ([]*AggregateMetric, string, error) {
	// newestTime: same as 'end'
	// duration: 0 if it should go forever
	// binTypes: "day", "hour", "minute", or "second"

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

	if len(cursor) > 0 {
		if c, err := datastore.DecodeCursor(cursor); err == nil {
			q = q.Start(c)
		} else {
			return []*AggregateMetric{}, "", err
		}
	}

	var aggregates []*AggregateMetric
	iter := q.Run(context)
	for limit < 0 || len(aggregates) < limit {
		var aggregate AggregateMetric
		if key, err := iter.Next(&aggregate); err == nil {
			aggregate.Key = key
		} else {
			break
		}
		aggregates = append(aggregates, &aggregate)
	}

	if len(aggregates) == limit {
		if c, err := iter.Cursor(); err == nil {
			cursor = c.String()
		}
	} else {
		cursor = ""
	}

	return aggregates, cursor, nil
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

	limit := int(-1) // no limit

	// parse optional limit
	if limitParam := request.FormValue("limit"); len(limitParam) > 0 {
		if limit64, err := strconv.ParseInt(limitParam, 10, 0); err != nil {
			context.Errorf("Error parsing limit: %s:%v", limitParam, err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		} else {
			limit = int(limit64)
		}
	}

	cursor := request.FormValue("cursor")

	// read aggregates
	aggregates, cursor, err := readAggregates(context, boardKey, namespace, binType, end, duration, limit, cursor)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	aggDtoList := makeAggregateDtoList(aggregates)
	response := JsonResponse{
		"results": aggDtoList,
	}
	if len(cursor) > 0 {
		url := request.URL
		values := url.Query()
		values.Set("cursor", cursor)
		values.Set("limit", strconv.Itoa(int(limit)))
		url.RawQuery = values.Encode()
		response["next"] = AbsURL(*url, request)
	}

	response.Write(writer)
}
