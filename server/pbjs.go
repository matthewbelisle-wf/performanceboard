package server

import (
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
)

func servePBJS(writer http.ResponseWriter, request *http.Request) {
	boardKey := mux.Vars(request)["board"]
	url, _ := router.Get("board").URL("board", boardKey)
	postURL := AbsURL(*url, request)
	display_data := make(map[string]string)
	display_data["post_url"] = postURL

	templates := template.Must(template.ParseFiles("server/templates/pb.js"))
	templates.ExecuteTemplate(writer, "pb.js", display_data)
	writer.Header().Set("content-type", "application/javascript")
}
