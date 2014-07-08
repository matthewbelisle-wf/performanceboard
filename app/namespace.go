package performanceboard

import (
	"encoding/json"
	"net/http"
)

type Namespace struct {
	Board    *Board
	Name     string
	Children Namespaces
}

type Namespaces []*Namespace

func (n Namespace) MarshalJSON() ([]byte, error) {
	api, _ := router.Get("namespace").URL("board", n.Board.Key.Encode(), "namespace", n.Name)
	return json.Marshal(Json{
		"name":     n.Name,
		"children": n.Children,
		"api":      AbsURL(api, n.Board.Context.Request().(*http.Request)), // TODO: Context.Request() is internal only?
	})
}
