package performanceboard

type Namespace struct {
	Name     string      `json:"name"`
	Children []*Namespace `json:"children"`
}

func getSubspaces(namespaces []*Namespace, top string, depth int64) []string {
	subspaces := []string{}
	for _, namespace := range namespaces {
		subspace := top + "." + namespace.Name
		subspaces = append(subspaces, subspace)
		more := getSubspaces(namespace.Children, subspace, depth - 1)
		subspaces = append(subspaces, more...)
	}
	return subspaces
}
