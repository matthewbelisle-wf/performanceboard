package performanceboard

import (
	"encoding/json"
	"net/http"
)

type Json map[string]interface{}

func (r Json) Write(w http.ResponseWriter) error {
	return WriteJson(r, w)
}

func WriteJson(v interface{}, w http.ResponseWriter) error {
	w.Header().Set("content-type", "application/json")
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	w.Write(bytes)
	return nil
}
