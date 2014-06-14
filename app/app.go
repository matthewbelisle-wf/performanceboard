// Sets up a REST interface
//
// Create = POST
// Read   = GET
// Update = PUT
// Delete = DELETE

package performanceboard

import (
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

var router *mux.Router

func init() {
	router = mux.NewRouter()
	router.HandleFunc("/api/{board}", createPost).Name("createPost").Methods("POST")
	router.HandleFunc("/api/", createBoard).Name("createBoard").Methods("POST")
	router.HandleFunc("/{client:.*}", client).Name("client")
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
