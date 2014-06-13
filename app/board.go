package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"net/http"
)

const BoardKind = "Board"

type Board struct {
	UserID string
}

func createBoard(writer http.ResponseWriter, request *http.Request) {
	context := appengine.NewContext(request)
	board := Board{}
	if u := user.Current(context); u != nil {
		board.UserID = u.ID
	}
	key, _ := datastore.Put(context, datastore.NewIncompleteKey(context, BoardKind, nil), &board)
	api, _ := router.Get("createPost").URL("board", key.Encode())
	JsonResponse{
		"board": key.Encode(),
		"api":   AbsURL(*api, request),
	}.Write(writer)
}
