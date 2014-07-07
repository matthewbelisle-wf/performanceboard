package performanceboard

import (
	// "appengine"
	// "appengine/datastore"
	// "encoding/json"
	// "github.com/gorilla/mux"
	// "net/http"
	// "time"
)

// func makeAggregateDtoList(metrics []AggregateMetric) []JsonResponse {
// 	aggDtoList := []JsonResponse{}
// 	for _, metric := range metrics {
// 		aggDTO := make(JsonResponse)
// 		aggDTO["start"] = metric.StartTime
// 		aggDTO["min"] = metric.Min
// 		aggDTO["max"] = metric.Max
// 		aggDTO["mean"] = metric.Mean
// 		aggDTO["count"] = metric.Count
// 		aggDtoList = append(aggDtoList, aggDTO)
// 	}
// 	return aggDtoList
// }

// func readAggregates(context appengine.Context,
// 	boardKey *datastore.Key, namespace string, binType string,
// 	newestTime time.Time, duration time.Duration,
// 	limit int, cursor string) ([]AggregateMetric, error) {
// 	q := datastore.NewQuery(AggregateMetricKind).
// 		Filter("Namespace =", namespace).
// 		Filter("BinType =", binType).
// 		Filter("StartTime <=", newestTime).
// 		Order("-StartTime").
// 		Ancestor(boardKey)

// 	if duration > 0 {
// 		oldestTime := newestTime.Add(-duration)
// 		q = q.Filter("StartTime >=", oldestTime)
// 	}

// 	//TODO use a limit and return a cursor
// 	var aggregates []AggregateMetric
// 	if _, err := q.GetAll(context, &aggregates); err != nil {
// 		context.Errorf("Error reading aggregates: %v", err)
// 		return nil, err
// 	}
// 	return aggregates, nil
// }

// // HTTP handlers

// func getAggregates(writer http.ResponseWriter, request *http.Request) {
// 	context := appengine.NewContext(request)
// 	encodedKey := mux.Vars(request)["board"]
// 	boardKey, err := datastore.DecodeKey(encodedKey)
// 	if err != nil {
// 		http.Error(writer, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	namespace := mux.Vars(request)["namespace"]
// 	binType := mux.Vars(request)["bin_type"]

// 	// TODO parse begin and end time filters from url form parameters
// 	// read aggregates
// 	aggregates, err := readAggregates(context, boardKey, namespace, binType, time.Now(), 0, 0, "")
// 	if err != nil {
// 		http.Error(writer, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	aggDtoList := makeAggregateDtoList(aggregates)
// 	b, err := json.Marshal(aggDtoList)
// 	if err != nil {
// 		http.Error(writer, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
// 	writer.Write(b)
// }
