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
	t := datastore.NewQuery(MetricKind).Ancestor(key).KeysOnly().Run(c)

	// Deletes metrics concurrently with 20 goroutines
	keyChan := make(chan *datastore.Key)
	errChan := make(chan error, 20)
	numErrs := 1
	go func() {
		for {
			key, err := t.Next(nil)
			if err == datastore.Done {
				break
			} else if err != nil {
				errChan <- err
				break
			}
			keyChan <- key
		}
		close(keyChan)
		errChan <- nil
	}()
	for key := range keyChan {
		go func(key *datastore.Key) {
			errChan <- datastore.Delete(c, key)
		}(key)
		numErrs++
	}
	for i := 0; i < numErrs; i++ {
		if err := <- errChan; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	board.ServeHTTP(w, r)
}
