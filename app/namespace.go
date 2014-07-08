package performanceboard

type Namespace struct {
	Name     string      `json:"name"`
	Children []Namespace `json:"children"`
}
