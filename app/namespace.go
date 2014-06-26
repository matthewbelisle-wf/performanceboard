package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"github.com/gorilla/mux"
	"net/http"
)

type Namespace struct {
	Name string `json:"name"`
	Api       string `json:"api"`
}

func getNamespaces(writer http.ResponseWriter, request *http.Request) {
	context := appengine.NewContext(request)
	boardKeyString := mux.Vars(request)["board"]
	q := datastore.NewQuery(TaxonomyKind).
		Filter("BoardKey =", boardKeyString)

	taxonomies := []Taxonomy{}
	if _, err := q.GetAll(context, &taxonomies); err != nil {
		context.Errorf("Error fetching Taxonomies:%v", err)
	}

	namespaces := []Namespace{}
	route := router.Get("namespace")
	for _, taxonomy := range taxonomies {
		url, _ := route.URL("board", boardKeyString, "namespace", taxonomy.Childspace)
		namespace := Namespace{Name: taxonomy.Childspace, Api: AbsURL(*url, request)}
		namespaces = append(namespaces, namespace)
	}

	JsonResponse{
		"namespaces": namespaces,
	}.Write(writer)
}