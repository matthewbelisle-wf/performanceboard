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
	Key     *datastore.Key    `datastore:"-" json:"key"`
	Context appengine.Context `datastore:"-" json:"-"`
	UserID  string            `datastore:"user_id" json:"-"`
}

func (b *Board) Json() (*Json, error) {
	namespaces, err := b.Namespaces()
	if err != nil {
		return nil, err
	}
	return &Json{
		"board":      b.Key.Encode(),
		"namespaces": namespaces,
	}, nil
}

func (b *Board) Namespaces() (Namespaces, error) {
	q := datastore.NewQuery(MetricKind).
		Ancestor(b.Key).
		Project("namespace").
		Distinct()
	hierarchy := map[string]*Namespace{} // {metricKeyString: *namespace}
	var err error
	for t := q.Run(b.Context); ; {
		metric := Metric{Context: b.Context}
		metric.Key, err = t.Next(&metric)
		if err == datastore.Done {
			break
		} else if err != nil {
			return nil, err
		}
		hierarchy[metric.Key.Encode()] = &Namespace{
			Name: metric.Namespace,
		}
	}
	namespaces := Namespaces{}
	for metricKeyString, namespace := range hierarchy {
		key, _ := datastore.DecodeKey(metricKeyString)
		if parentKey := key.Parent(); parentKey.Kind() == MetricKind {
			parent, _ := hierarchy[parentKey.Encode()]
			parent.Children = append(parent.Children, namespace)
		} else {
			namespaces = append(namespaces, namespace)
		}
	}
	return namespaces, nil
}

// HTTP handlers

func getBoard(w http.ResponseWriter, r *http.Request) {
	keyString := mux.Vars(r)["board"]
	key, err := datastore.DecodeKey(keyString)
	if err != nil {
		http.Error(w, "Invalid key string: %s"+keyString, http.StatusBadRequest)
		return
	}
	c := appengine.NewContext(r)
	board := Board{Context: c, Key: key}
	if err = datastore.Get(c, key, &board); err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	json, err := board.Json()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	json.Write(w)
}

func createBoard(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	board := Board{Context: c}
	if u := user.Current(c); u != nil {
		board.UserID = u.ID
	}
	var err error
	board.Key, err = datastore.Put(c, datastore.NewIncompleteKey(c, BoardKind, nil), &board)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	json, err := board.Json()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	json.Write(w)
}
