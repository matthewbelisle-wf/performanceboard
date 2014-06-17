package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"time"
)

type Metric struct {
	DsKey *datastore.Key `datastore:"-"`
	Key   string
	Start time.Time
	End   time.Time
}

// HTTP handlers

func getMetrics(writer http.ResponseWriter, request *http.Request) {
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

func postMetrics(writer http.ResponseWriter, request *http.Request) {
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
}
