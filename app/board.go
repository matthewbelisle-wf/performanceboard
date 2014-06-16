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

func (board *Board) WriteJson(writer http.ResponseWriter, request *http.Request) {
	api, _ := router.Get("board").URL("board", board.Key.Encode())
	JsonResponse{
		"board": board.Key.Encode(),
		"api":   AbsURL(*api, request),
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
	board.WriteJson(writer, request)
}

func createBoard(writer http.ResponseWriter, request *http.Request) {
	context := appengine.NewContext(request)
	board := Board{}
	if u := user.Current(context); u != nil {
		board.UserID = u.ID
	}
	key, _ := datastore.Put(context, datastore.NewIncompleteKey(context, BoardKind, nil), &board)
	board.Key = key
	board.WriteJson(writer, request)
}
