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
	UserID string `datastore:"user_id"`
}

func (b *Board) Json(c appengine.Context, key *datastore.Key) (*Json, error) {
	namespaces, err := b.Namespaces(c, key)
	if err != nil {
		return nil, err
	}
	return &Json{
		"board":      key.Encode(),
		"namespaces": namespaces,
	}, nil
}

func (b *Board) Namespaces(c appengine.Context, key *datastore.Key) (Namespaces, error) {
	q := datastore.NewQuery(MetricKind).
		Ancestor(key).
		Project("namespace").
		Distinct()
	hierarchy := map[string]*Namespace{} // {metricKeyString: *namespace}
	for t := q.Run(c); ; {
		metric := Metric{}
		metricKey, err := t.Next(&metric)
		if err == datastore.Done {
			break
		} else if err != nil {
			return nil, err
		}
		hierarchy[metricKey.Encode()] = &Namespace{
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
	board := Board{}
	if err = datastore.Get(c, key, &board); err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	json, err := board.Json(c, key)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	json.Write(w)
}

func createBoard(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	board := Board{}
	if u := user.Current(c); u != nil {
		board.UserID = u.ID
	}
	key, err := datastore.Put(c, datastore.NewIncompleteKey(c, BoardKind, nil), &board)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	json, err := board.Json(c, key)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	json.Write(w)
}
