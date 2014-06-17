package performanceboard

import (
	"appengine/datastore"
	"time"
)

type Metric struct {
	DsKey *datastore.Key `datastore:"-"`
	Key   string
	Start time.Time
	End   time.Time
}
