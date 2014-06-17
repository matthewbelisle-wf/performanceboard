package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"time"
)

const PostKind = "Post"

type Post struct {
	Key       *datastore.Key `datastore:"-"`
	Body      string         `datastore:",noindex"`
	Timestamp time.Time      `datastore:",noindex"`
}

func (post *Post) WriteJson(writer http.ResponseWriter, request *http.Request) {
	boardKey := post.Key.Parent()
	api, _ := router.Get("board").URL("board", boardKey.Encode())
	JsonResponse{
		"api": AbsURL(*api, request),
		"post": post.Key.Encode(),
		"board": boardKey.Encode(),
	}.Write(writer)
}

func createPost(writer http.ResponseWriter, request *http.Request) {
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
	post.WriteJson(writer, request)
}
