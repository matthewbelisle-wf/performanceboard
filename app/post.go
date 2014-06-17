package performanceboard

import (
	"appengine/datastore"
	"time"
)

const PostKind = "Post"

type Post struct {
	Key       *datastore.Key `datastore:"-"`
	Body      string         `datastore:",noindex"`
	Timestamp time.Time      `datastore:",noindex"`
}
