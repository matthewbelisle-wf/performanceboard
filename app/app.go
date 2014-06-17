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
	router.HandleFunc("/api/{board}", getMetrics).Methods("GET").Name("board")
	router.HandleFunc("/api/{board}", postMetrics).Methods("POST")
	router.HandleFunc("/api/{board}", methodNotAllowed)
	router.HandleFunc("/api/", createBoard).Methods("POST")
	router.HandleFunc("/api/", methodNotAllowed)
	router.HandleFunc("/{client:.*}", client)
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

func methodNotAllowed(writer http.ResponseWriter, request *http.Request) {
	http.Error(writer, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}
