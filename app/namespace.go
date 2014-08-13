package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"github.com/gorilla/mux"
	"net/http"
)

type Namespace struct {
	Name     string       `json:"name"`
	Api      string       `json:"api"`
	Children []*Namespace `datastore:"-" json:"children"`
}

func readNamespaceChildren(context appengine.Context, boardKeyString string, rootNamespace string) []string {
	q := datastore.NewQuery(TaxonomyKind).
		Filter("BoardKey =", boardKeyString).
		Filter("Namespace =", rootNamespace)

	taxonomies := []Taxonomy{}
	if _, err := q.GetAll(context, &taxonomies); err != nil {
		context.Errorf("Error fetching Taxonomies:%v", err)
	}
	names := []string{}
	for _, taxonomy := range taxonomies {
		names = append(names, taxonomy.Childspace)
	}
	return names
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

	hierarchy := map[string]*Namespace{}
	parents := map[string]string{}
	namespaces := []*Namespace{} // Top level namespaces
	route := router.Get("namespace")
	for _, taxonomy := range taxonomies {
		url, _ := route.URL("board", boardKeyString, "namespace", taxonomy.Childspace)
		hierarchy[taxonomy.Childspace] = &Namespace{
			Name: taxonomy.Childspace,
			Api:  AbsURL(*url, request),
		}
		parents[taxonomy.Childspace] = taxonomy.Namespace
	}
	for _, namespace := range hierarchy {
		if parent, ok := hierarchy[parents[namespace.Name]]; ok {
			parent.Children = append(parent.Children, namespace)
		} else {
			namespaces = append(namespaces, namespace)
		}
	}

	JsonResponse{
		"namespaces": namespaces,
	}.Write(writer)
}
