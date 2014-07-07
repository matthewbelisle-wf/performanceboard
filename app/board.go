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
	boardKeyString := mux.Vars(r)["board"]
	boardKey, err := datastore.DecodeKey(boardKeyString)
	if err != nil {
		http.Error(w, "Invalid boardKey: %s"+boardKeyString, http.StatusBadRequest)
	}
	q := datastore.NewQuery(MetricKind).
		Ancestor(boardKey).
		Project("namespace").
		Distinct()
	c := appengine.NewContext(r)
	namespaces := map[string]*Namespace{} // {metricKeyString: *namespace}
	topNamespaces := []*Namespace{}
	for t := q.Run(c); ; {
		metric := Metric{}
		metricKey, err := t.Next(&metric)
		if err == datastore.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		api, _ := router.Get("namespace").URL("board", boardKey.Encode(), "namespace", metric.Namespace)
		namespaces[metricKey.Encode()] = &Namespace{
			Name: metricKey.StringID(),
			Api:  AbsURL(*api, r),
		}
	}
	for metricKeyString, namespace := range namespaces {
		key, _ := datastore.DecodeKey(metricKeyString)
		if parentKey := key.Parent(); parentKey.Kind() == MetricKind {
			parent, _ := namespaces[parentKey.Encode()]
			parent.Children = append(parent.Children, *namespace)
		} else {
			topNamespaces = append(topNamespaces, namespace)
		}
	}
	JsonResponse{
		"namespaces": topNamespaces,
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
