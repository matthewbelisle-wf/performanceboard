// Sets up a REST interface
//
// Create = POST
// Read   = GET
// Update = PUT
// Delete = DELETE

package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

var router *mux.Router

func init() {
	router = mux.NewRouter()
	router.HandleFunc("/api/{board}", createPost).Name("createPost").Methods("POST")
	router.HandleFunc("/api/", createBoard).Name("createBoard").Methods("POST")
	router.HandleFunc("/", client)
	http.Handle("/", router)
}

var indexHtml, _ = ioutil.ReadFile("static/index.html")

func client(writer http.ResponseWriter, request *http.Request) {
	if !Authorized(writer, request) {
		return
	}
	writer.Header().Set("content-type", "text/html")
	writer.Write(indexHtml)
}

func createBoard(writer http.ResponseWriter, request *http.Request) {
	context := appengine.NewContext(request)
	board := Board{}
	key := CreateBoard(&board, context)
	api, _ := router.Get("createPost").URL("board", key.Encode())
	JsonResponse{
		"board": key.Encode(),
		"api": AbsURL(*api, request),
	}.Write(writer)
}

func createPost(writer http.ResponseWriter, request *http.Request) {
	// Checks that the key is valid
	encodedBoardKey := mux.Vars(request)["board"]
	boardKey, err := datastore.DecodeKey(encodedBoardKey)
	if err != nil || boardKey.Kind() != BoardKind {
		http.Error(writer, "Invalid key: "+encodedBoardKey, http.StatusBadRequest)
		return
	}

	defer request.Body.Close()
	body, _ := ioutil.ReadAll(request.Body)
	post := Post{
		BoardKey: boardKey,
		Body:     string(body),
	}
	context := appengine.NewContext(request)
	key := CreatePost(&post, context)
	JsonResponse{
		"key":       key.Encode(),
		"board":     AbsURL(*request.URL, request),
		"body":      post.Body,
		"timestamp": post.Timestamp.Unix(),
	}.Write(writer)
}
