package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"github.com/gorilla/mux"
	"net/http"
)

const BoardKind = "Board"

type Board struct {
	Key    *datastore.Key `datastore:"-"`
	UserID string
}

func (board *Board) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	api, _ := router.Get("board").URL("board", board.Key.Encode())
	JsonResponse{
		"board": board.Key.Encode(),
		"api":   AbsURL(*api, request),
	}.Write(writer)
}

// HTTP handlers

func getBoard(w http.ResponseWriter, r *http.Request) {
	keyString := mux.Vars(r)["board"]
	key, err := datastore.DecodeKey(keyString)
	if err != nil {
		http.Error(w, "Invalid key: %s"+keyString, http.StatusBadRequest)
	}
	q := datastore.NewQuery(MetricKind).
		Ancestor(key).
		Project("namespace").
		Distinct()
	c := appengine.NewContext(r)
	metrics := map[*datastore.Key]Metric{}
	topMetrics := []Metric{}
	for t := q.Run(c);; {
		metric := Metric{}
		metricKey, err := t.Next(&metric)
		if err == datastore.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		metrics[metricKey] = metric
	}
	for metricKey, metric := range metrics {
		if parentKey := metricKey.Parent(); parentKey.Kind() == MetricKind {
			parent := metrics[parentKey]
			parent.Children = append(parent.Children, metric)
		} else {
			topMetrics = append(topMetrics, metric)
		}
	}
	JsonResponse{
		"metrics": topMetrics,
	}.Write(w)
}

func createBoard(writer http.ResponseWriter, request *http.Request) {
	context := appengine.NewContext(request)
	board := Board{}
	if u := user.Current(context); u != nil {
		board.UserID = u.ID
	}
	board.Key, _ = datastore.Put(context, datastore.NewIncompleteKey(context, BoardKind, nil), &board)
	board.ServeHTTP(writer, request)
}
