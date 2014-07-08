package performanceboard

type Namespace struct {
	Name     string     `json:"name"`
	Children Namespaces `json:"children"`
}

type Namespaces []*Namespace
