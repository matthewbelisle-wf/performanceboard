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
	Namespace string `json:"namespace"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Meta      map[string]interface{} `json:"meta"`
	Children  []PostBody `json:"children"`
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
	metric.Namespace = data.Namespace
	if ancestor != nil {
		metric.Namespace = ancestor.Namespace + metric.Namespace
	}
	meta, err := json.Marshal(data.Meta)
	if err != nil {
		c.Infof("failed to marshall metadata: %v", err)
	}
	metric.Meta = string(meta)
	metric.Start = data.Start
	metric.End = data.End

	// compose an idempotent key for this data (allows Post data to be digested repeatedly)
	keyID := fmt.Sprintf("%s:%s:%v:%v", postKey, metric.Namespace, data.Start, data.End)
	if ancestor == nil {
		metric.Key = datastore.NewKey(c, MetricKind, keyID, 0, nil)
	} else {
		metric.Key = datastore.NewKey(c, MetricKind, keyID, 0, ancestor.Key)
	}
	_, err = datastore.Put(c, metric.Key, &metric)
    if err != nil {
        c.Infof("Error on Put:%v", err)
        return
    }
	// recursive storage for child objects to the same entity table
	for _, child := range data.Children {
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
        c.Infof("Error in Unmarshal:%v", err)
		return
	}

    c.Infof("namespace:%s", body.Namespace)
    c.Infof("body:%v", body)
    c.Infof("post:%v", post)
    c.Infof("postBody:%v", post.Body)

	storeMetric(c, postKey, body, nil)
})
