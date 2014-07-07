package performanceboard

type Namespace struct {
	Name     string      `json:"name"`
	Api      string      `json:"api"`
	Children []Namespace `json:"children"`
}
