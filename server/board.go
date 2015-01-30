package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const BoardKind = "Board"

type Board struct {
	Key    *datastore.Key `datastore:"-"`
	Name   string
	UserID string
}

func (board *Board) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	api, _ := router.Get("board").URL("board", board.Key.Encode())
	JsonResponse{
		"board": board.Key.Encode(),
		"api":   AbsURL(*api, request),
	}.Write(writer)
}

func createBoard(context appengine.Context, writer http.ResponseWriter, request *http.Request) {
	boardName := request.FormValue("name")
	board := Board{}
	if u := user.Current(context); u != nil {
		board.UserID = u.ID
		board.Name = boardName
	}
	board.Key, _ = datastore.Put(context, datastore.NewIncompleteKey(context, BoardKind, nil), &board)
	board.ServeHTTP(writer, request)
}

func listBoards(context appengine.Context, writer http.ResponseWriter, request *http.Request) {
	q := datastore.NewQuery(BoardKind).KeysOnly()
	if keys, err := q.GetAll(context, nil); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		keyList := []JsonResponse{}
		log.Println("keylist:", keyList)
		for _, key := range keys {
			api, _ := router.Get("client").URL("client", key.Encode())
			log.Println("api:", api)

			keyList = append(keyList, JsonResponse{
				"name": "BOARD",
				"url":  AbsURL(*api, request),
			})
		}
		JsonResponse{
			"results": keyList,
		}.Write(writer)
	}
}

func clearBoard(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	keyString := mux.Vars(r)["board"]
	key, err := datastore.DecodeKey(keyString)
	if err != nil || key.Kind() != BoardKind {
		http.Error(w, "Invalid board key: "+keyString, http.StatusBadRequest)
		return
	}

	board := Board{Key: key}
	if err = datastore.Get(c, key, &board); err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	q := datastore.NewQuery(MetricKind).Ancestor(key).KeysOnly().Limit(2000)
	for {
		if keys, err := q.GetAll(c, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if len(keys) == 0 {
			break
		} else if err = datastore.DeleteMulti(c, keys); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	board.ServeHTTP(w, r)
}
