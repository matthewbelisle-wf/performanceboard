// Sets up a REST interface
//
// Create = POST
// Read   = GET
// Update = PUT
// Delete = DELETE

package server

import (
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

var router *mux.Router

func Init() {
	router = mux.NewRouter()
	router.HandleFunc("/api/post/{post_key}", getPost).Methods("GET").Name("get_post")
	router.HandleFunc("/api/post", methodNotAllowed)
	router.HandleFunc("/api/{board}/{namespace}/{bin_type}", getAggregates).Methods("GET").Name("aggregate")
	router.HandleFunc("/api/{board}/{namespace}/{bin_type}", methodNotAllowed)
	router.HandleFunc("/api/{board}/{namespace}", getMetrics).Methods("GET").Name("namespace")
	router.HandleFunc("/api/{board}/{namespace}", methodNotAllowed)
	router.HandleFunc("/api/{board}", getNamespaces).Methods("GET").Name("board")
	router.HandleFunc("/api/{board}", postMetric).Methods("POST")
	router.HandleFunc("/api/{board}", clearBoard).Methods("PUT")
	router.HandleFunc("/api/{board}", methodNotAllowed)
	router.HandleFunc("/api/", createBoard).Methods("POST")
	router.HandleFunc("/api/", listBoards).Methods("GET")
	router.HandleFunc("/api/", methodNotAllowed)
	router.HandleFunc("/pbjs/{board}", servePBJS).Methods("GET")
	router.HandleFunc("/{client:.*}", client).Name("client")
	http.Handle("/", router)
}

var indexHtml, _ = ioutil.ReadFile("index.html")

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
