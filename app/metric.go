package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
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
		if len(metric.Children) > 0 {
			metricDTO["children"] = makeMetricDtoList(metric.Children)
		}
		metricDtoList = append(metricDtoList, metricDTO)
	}
	return metricDtoList
}

func readMetricChildren(context appengine.Context, boardKeyString string, parent *Metric, depth int) ([]Metric, error) {
	childNamespaces := readNamespaceChildren(context, boardKeyString, parent.Namespace)
	var allChildren []Metric
	for _, namespace := range childNamespaces {
		q := datastore.NewQuery(MetricKind).
			Filter("Namespace =", namespace).
			Ancestor(parent.Key)

		children := []Metric{}
		keys, err := q.GetAll(context, &children)
		if err != nil {
			return nil, err
		}
		if depth != 0 {
			for i, child := range children {
				child.Key = keys[i]
				if grandChildren, err := readMetricChildren(context, boardKeyString, &child, (depth - 1)); err != nil {
					context.Errorf("error reading children:%v", err)
					return nil, err
				} else {
					if len(grandChildren) > 0 {
						child.Children = grandChildren
					}
				}
			}
		}
		for _, child := range children {
			allChildren = append(allChildren, child)
		}
	}

	return allChildren, nil
}

func readMetrics(context appengine.Context, boardKey *datastore.Key, namespace string,
	newestTime time.Time, duration time.Duration, depth int) ([]Metric, error) {
	// Duration is used instead of oldestTime to help support unbound queries

	q := datastore.NewQuery(MetricKind).
		Filter("Namespace =", namespace).
		Filter("Start <=", newestTime).
		Order("-Start").
		Ancestor(boardKey)

	if duration > 0 {
		oldestTime := newestTime.Add(-duration)
		q = q.Filter("Start >=", oldestTime)
	}

	// TODO use a limit and return a cursor
	var metrics []Metric
	keys, err := q.GetAll(context, &metrics)
	if err != nil {
		context.Errorf("Error reading metrics: %v", err)
		return nil, err
	}

	// foreach Metric recursively fetch children
	if depth != 0 {
		for i, metric := range metrics {
			metric.Key = keys[i]
			if kids, err := readMetricChildren(context, boardKey.Encode(), &metric, (depth - 1)); err != nil {
				context.Errorf("Error reading top-level Metrics children for %v:%v", metric, err)
				return nil, err
			} else {
				if len(kids) > 0 {
					metrics[i].Children = kids
				}
			}
		}
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

	// evaluate newest point in request timeline
	end := time.Now()
	if endParam := request.FormValue("end"); len(endParam) > 0 {
		if end, err = time.Parse(time.RFC3339, endParam); err != nil {
			context.Errorf("Error parsing end:%s:%v", endParam, err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}
	}

	// evalutate oldest point in request timeline
	duration := time.Duration(0)
	if startParam := request.FormValue("start"); len(startParam) > 0 {
		if start, err := time.Parse(time.RFC3339, startParam); err != nil {
			context.Errorf("Error parsing start:%s:%v", startParam, err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		} else {
			duration = end.Sub(start)
		}
	}

	depth := int64(0)
	if depthParam := request.FormValue("depth"); len(depthParam) > 0 {
		if depth, err = strconv.ParseInt(depthParam, 10, 0); err != nil {
			context.Errorf("Error parsing depth: %s:%v", depthParam, err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
	}

	metrics, err := readMetrics(context, boardKey, namespace, end, duration, int(depth))
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
