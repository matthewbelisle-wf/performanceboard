package performanceboard

import (
	"encoding/json"
	"net/http"
)

type JsonResponse map[string]interface{}

func (response JsonResponse) Write(writer http.ResponseWriter) {
	writer.Header().Set("content-type", "application/json")
	bytes, _ := json.MarshalIndent(response, "", "  ")
	writer.Write(bytes)
	writer.Write([]byte{'\n'})
}
