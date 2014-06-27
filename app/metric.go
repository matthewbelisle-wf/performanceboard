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

func makeMetricDtoList(metrics []Metric) []JsonResponse {
	metricDtoList := []JsonResponse{}
	for _, metric := range metrics {
		metricDTO := make(JsonResponse)
		metricDTO["start"] = metric.Start
		metricDTO["end"] = metric.End
		if len(metric.Meta) > 0 {
			//conditionally add the metric data (helps trim the network traffic)
			var meta map[string]interface{}
			if err := json.Unmarshal([]byte(metric.Meta), &meta); err == nil {
				metricDTO["meta"] = meta
			}
		}
		metricDtoList = append(metricDtoList, metricDTO)
	}
	return metricDtoList
}

func readMetrics(context appengine.Context,
	boardKey *datastore.Key, namespace string,
	newestTime time.Time, duration time.Duration) ([]Metric, error) {
	q := datastore.NewQuery(MetricKind).
		Filter("Namespace =", namespace).
		Filter("Start <", newestTime).
		Order("-Start").
		Ancestor(boardKey)

	if duration > 0 {
		oldestTime := newestTime.Add(-duration)
		q = q.Filter("Start >", oldestTime)
	}

	var metrics []Metric
	//TODO use a limit and return a cursor
	if _, err := q.GetAll(context, &metrics); err != nil {
		context.Infof("Error reading metrics: %v", err)
		return nil, err
	}
	return metrics, nil
}

// HTTP handlers

func getMetrics(writer http.ResponseWriter, request *http.Request) {
	context := appengine.NewContext(request)
	encodedKey := mux.Vars(request)["board"]
	boardKey, err := datastore.DecodeKey(encodedKey)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	namespace := mux.Vars(request)["namespace"]
	var duration time.Duration
	if param := request.FormValue("start"); len(param) > 0 {
		start, err := time.Parse(time.RFC3339, param)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}
		duration = time.Now().Sub(start)
	}	
	metrics, err := readMetrics(context, boardKey, namespace, time.Now(), duration)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	metricsDtoList := makeMetricDtoList(metrics)
	b, err := json.Marshal(metricsDtoList)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(b)
}

// TODO memcache the board entity and validate boardKey against it
func postMetric(writer http.ResponseWriter, request *http.Request) {
	// Checks that the key is valid
	encodedBoardKey := mux.Vars(request)["board"]
	boardKey, err := datastore.DecodeKey(encodedBoardKey)
	if err != nil || boardKey.Kind() != BoardKind {
		http.Error(writer, "Invalid Board key: "+encodedBoardKey, http.StatusBadRequest)
		return
	}
	defer request.Body.Close()
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	post := Post{
		Body:      string(body),
		Timestamp: time.Now(),
	}
	context := appengine.NewContext(request)
	post.Key, err = datastore.Put(context, datastore.NewIncompleteKey(context, PostKind, boardKey), &post)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	digestPost.Call(context, post.Key.Encode())
}
