package performanceboard

type Namespace struct {
	Name     string     `json:"name"`
	Children Namespaces `json:"children"`
}

type Namespaces []*Namespace

// Recursively searches namespaces for the children of a given namespace
func (n Namespaces) Childspaces(namespace string) Namespaces {
	var search func(needle string, haystack Namespaces) Namespaces
	search = func(needle string, haystack Namespaces) Namespaces {
		if needle == namespace {
			return haystack
		}
		for _, child := range haystack {
			needle2 := needle + "." + child.Name
			if found := search(needle2, child.Children); found != nil {
				return found
			}
		}
		return nil
	}
	return search(namespace, n)
}
