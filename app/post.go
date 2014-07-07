package performanceboard

import (
	"time"
)

const PostKind = "Post"

type Post struct {
	Body      string    `datastore:",noindex"`
	Timestamp time.Time `datastore:",noindex"`
}
