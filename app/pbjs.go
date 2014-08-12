package performanceboard

import (
    // "appengine"
	// "github.com/gorilla/mux"
    "html/template"
	"net/http"
	"net/url"
)



func servePBJS(writer http.ResponseWriter, request *http.Request) {
	// context := appengine.NewContext(request)

	boardKey := request.FormValue("post_key")

	boardUrl := "/api/" + boardKey
	url, err := url.Parse(boardUrl)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	postURL := AbsURL(*url, request)
    // context.Infof("postUrl:", postURL)
    display_data := make(map[string]string)
	display_data["post_url"] = postURL

    templates := template.Must(template.ParseFiles("static/pb.js"))
	templates.ExecuteTemplate(writer, "pb.js", display_data)
}
