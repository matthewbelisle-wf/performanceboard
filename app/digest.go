package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"appengine/delay"
)

var digestPost = delay.Func("key", func(c appengine.Context, key datastore.Key) {
	c.Infof("digestPost: %v", key)
})
