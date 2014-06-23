package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"github.com/gorilla/mux"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"time"
)

type MetricDTO struct {
	Start time.Time `json:"start"`
	End time.Time `json:"end"`
	Meta map[string]interface{} `json:"meta"`
}

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
	q := datastore.NewQuery(MetricKind).
                Filter("Namespace =", namespace).
                Order("-Start").
                Ancestor(boardKey)

	var metrics []Metric
    _, err = q.GetAll(context, &metrics)
    metricsDtoList := makeMetricDtoList(metrics)
    b, err := json.Marshal(metricsDtoList)
    if err != nil {
        return
    }
    writer.Header().Set("Content-Type", "application/json; charset=utf-8")
    writer.Write(b)
}


func getNamespaces(writer http.ResponseWriter, request *http.Request) {
	context := appengine.NewContext(request)
	boardKeyString := mux.Vars(request)["board"]
	context.Infof("looking up board %s", boardKeyString)
	q := datastore.NewQuery(TaxonomyKind).
		Filter("BoardKey =", boardKeyString)

    var taxonomies []Taxonomy
    if _, err := q.GetAll(context, &taxonomies); err != nil {
    	context.Errorf("Error fetching Taxonomies:%v", err)
    }

    context.Infof("Found %d entries", len(taxonomies))
    namespaces := []string{}
    route := router.Get("namespace")
    for _, taxonomy := range taxonomies {
    	url, _ := route.URL("board", boardKeyString, "namespace", taxonomy.Childspace)
		namespaces = append(namespaces, AbsURL(*url, request))
    }

	JsonResponse{
		"series": namespaces,
	}.Write(writer)
}


func getBoard(writer http.ResponseWriter, request *http.Request) {
	encodedKey := mux.Vars(request)["board"]
	key, err := datastore.DecodeKey(encodedKey)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	context := appengine.NewContext(request)
	board := Board{Key: key}
	if err := datastore.Get(context, key, &board); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	board.ServeHTTP(writer, request)
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
