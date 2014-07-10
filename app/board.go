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
	keys := []*datastore.Key{}
	t := datastore.NewQuery(MetricKind).Ancestor(key).KeysOnly().Run(c)
	keyChan := make(chan *datastore.Key, 50)
	go func() {
		for {
			key, err := t.Next(nil)
			if err == datastore.Done {
				keyChan <- key
				break
			} else if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				break
			}
		}
		close(keyChan)
	}()
	errChan := make(chan error, 50)
	go func() {
		for key := range keyChan {
			go func(key *datastore.Key) {
				errChan <- datastore.Delete(c, key)
			}(key)
		}
		close(errChan)
	}()
	for err := range errChan {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	board.ServeHTTP(w, r)
}
