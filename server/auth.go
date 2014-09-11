// Authentication and Authorization
//
// Authentication is handled by Appengine itself.
// https://developers.google.com/appengine/docs/go/users/
//
// Data CREATEs require no authorization (createBoard, createPost, etc.)
// Data READs require authorization if the users.yaml file is populated

package server

import (
	"appengine"
	"appengine/user"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"net/http"
)

// Builds a map of authorized users (instead of a slice for constant time lookup)
var users = func() map[string]bool {
	bytes, _ := ioutil.ReadFile("users.yaml")
	var list []string
	if err := yaml.Unmarshal(bytes, &list); err != nil {
		panic("Invalid syntax in user.yaml: " + err.Error())
	}
	users := make(map[string]bool)
	for _, email := range list {
		users[email] = true
	}
	return users
}()

func Authenticated(writer http.ResponseWriter, request *http.Request) bool {
	context := appengine.NewContext(request)
	if user.Current(context) != nil {
		return true
	}
	url, _ := user.LoginURL(context, request.URL.String())
	writer.Header().Set("Location", url)
	writer.WriteHeader(http.StatusFound)
	return false
}

func Authorized(writer http.ResponseWriter, request *http.Request) bool {
	if len(users) == 0 {
		return true // Everyone is authorized
	}
	if !Authenticated(writer, request) {
		return false
	}
	context := appengine.NewContext(request)
	email := user.Current(context).Email
	if users[email] {
		return true
	}
	http.Error(writer, email+" not in users.yaml", http.StatusUnauthorized)
	return false
}
