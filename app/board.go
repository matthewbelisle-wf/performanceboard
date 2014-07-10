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

func createBoard(writer http.ResponseWriter, request *http.Request) {
	context := appengine.NewContext(request)
	board := Board{}
	if u := user.Current(context); u != nil {
		board.UserID = u.ID
	}
	board.Key, _ = datastore.Put(context, datastore.NewIncompleteKey(context, BoardKind, nil), &board)
	board.ServeHTTP(writer, request)
}

func clearBoard(w http.ResponseWriter, r *http.Request) {
	keyString := mux.Vars(r)["board"]
	key, err := datastore.DecodeKey(keyString)
	if err != nil || key.Kind() != BoardKind {
		http.Error(w, "Invalid board key: "+keyString, http.StatusBadRequest)
		return
	}
	c := appengine.NewContext(r)
	board := Board{Key: key}
	if err = datastore.Get(c, key, &board); err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	q := datastore.NewQuery(MetricKind).Ancestor(key).KeysOnly().Limit(2000)
	for {
		keys, err := q.GetAll(c, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(keys) == 0 {
			break
		}
		if err = datastore.DeleteMulti(c, keys); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return			
		}
	}

	board.ServeHTTP(w, r)
}
