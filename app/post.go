package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"time"
)

const PostKind = "Post"

type Post struct {
	BoardKey  *datastore.Key
	Body      string
	Timestamp time.Time
}

// TODO: post.Board()

func CreatePost(post *Post, context appengine.Context) *datastore.Key {
	post.Timestamp = time.Now()
	key, _ := datastore.Put(context, datastore.NewIncompleteKey(context, PostKind, nil), post)
	return key
}
