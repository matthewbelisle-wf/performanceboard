package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"time"
)

func readAggregates(context appengine.Context,
	boardKey *datastore.Key, namespace string,
	newestTime time.Time, duration time.Duration) ([]Metric, error) {
	q := datastore.NewQuery(AggregateMetricKind).
		Filter("Namespace =", namespace).
		Filter("StartTime <", newestTime).
		Order("-StartTime").
		Ancestor(boardKey)

	if duration > 0 {
		oldestTime := newestTime.Add(-duration)
		q = q.Filter("Start >", oldestTime)
	}

	var aggregates []AggregateMetric
	//TODO use a limit and return a cursor
	if _, err := q.GetAll(context, &aggregates); err != nil {
		context.Infof("Error reading aggregates: %v", err)
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
	// TODO parse begin and end times from url form parameters
	// TODO read aggregates
}
