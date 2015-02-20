package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"github.com/gorilla/mux"
	"net/http"
)

func getPost(writer http.ResponseWriter, request *http.Request) {
	context := appengine.NewContext(request)
	encodedKey := mux.Vars(request)["post_key"]
	postKey, err := datastore.DecodeKey(encodedKey)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	post := new(Post)
	if err := datastore.Get(context, postKey, post); err != nil {
		http.Error(writer, err.Error(), 404)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write([]byte(post.Body))
}
