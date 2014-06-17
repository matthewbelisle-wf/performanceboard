package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
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
