package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
)

const BoardKind = "Board"

type Board struct {
	UserID string
}

func CreateBoard(board *Board, context appengine.Context) *datastore.Key {
	if u := user.Current(context); u != nil {
		board.UserID = u.ID
	}
	key, _ := datastore.Put(context, datastore.NewIncompleteKey(context, BoardKind, nil), board)
	return key
}
