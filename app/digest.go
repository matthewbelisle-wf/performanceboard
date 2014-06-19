package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"appengine/delay"
	"encoding/json"
	"fmt"
	"time"
)

/* data package parsing structure, not stored to disk. */
type PostBody struct {
	namespace string
	start     time.Time
	end       time.Time
	meta      map[string]interface{}
	children  []PostBody
}

/* A Metric is a namespaced measurement. A metric's parent (for Ancestor queries)
is the metric that contained it in the origional Post. Therefore, a Metric
with the namespace 'X.Y.Z' will have an Ancestor Metric with a namespace
of 'X.Y'.
*/
const MetricKind = "Metric"

type Metric struct {
	Key       *datastore.Key `datastore:"-"`
	Namespace string         // dot seperated name hierarchy
	Meta      string         `datastore:",noindex"` // stringified JSON object
	Start     time.Time      // UTC
	End       time.Time      // UTC
}

func storeMetric(c appengine.Context, postKey string, data PostBody, ancestor *Metric) {
	metric := Metric{}
	metric.Namespace = data.namespace
	if ancestor != nil {
		metric.Namespace = ancestor.Namespace + metric.Namespace
	}
	meta, err := json.Marshal(data.meta)
	if err != nil {
		c.Infof("failed to marshall metadata: %v", err)
	}
	metric.Meta = string(meta)
	metric.Start = data.start
	metric.End = data.end

	// compose an idempotent key for this data (allows Post data to be digested repeatedly)
	keyID := fmt.Sprintf("%s:%s:%v:%v", postKey, metric.Namespace, data.start, data.end)
	if ancestor == nil {
		key := datastore.NewKey(c, MetricKind, keyID, 0, nil)
		_, err = datastore.Put(c, key, metric)
	} else {
		key := datastore.NewKey(c, MetricKind, keyID, 0, ancestor.Key)
		_, err = datastore.Put(c, key, metric)
	}

	// recursive storage for child objects to the same entity table
	for _, child := range data.children {
		storeMetric(c, postKey, child, &metric)
	}
}

var digestPost = delay.Func("key", func(c appengine.Context, postKey string) {
	post := Post{}
	key, _ := datastore.DecodeKey(postKey)
	if err := datastore.Get(c, key, &post); err != nil {
		return
	}
	body := PostBody{}
	if err := json.Unmarshal([]byte(post.Body), &body); err != nil {
		return
	}

	storeMetric(c, postKey, body, nil)
})
