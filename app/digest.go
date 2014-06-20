package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"appengine/delay"
	"encoding/json"
	"fmt"
	"time"
)

// data package parsing structure, not stored to disk.
type PostBody struct {
	Namespace string                 `json:"namespace"`
	Start     time.Time              `json:"start"`
	End       time.Time              `json:"end"`
	Meta      map[string]interface{} `json:"meta"`
	Children  []PostBody             `json:"children"`
}

// A Metric is a namespaced measurement. A metric's parent (for Ancestor queries)
// is the metric that contained it in the origional Post. Therefore, a Metric
// with the namespace 'X.Y.Z' will have an Ancestor Metric with a namespace
// of 'X.Y'.
const MetricKind = "Metric"

type Metric struct {
	Key       *datastore.Key `datastore:"-"`
	Namespace string         // dot seperated name hierarchy
	Meta      string         `datastore:",noindex"` // stringified JSON object
	Start     time.Time      // UTC
	End       time.Time      // UTC
}

// The Taxonomy table defines namespace relationships for fast lookup
// Given static assignment of a namespace to a measurement on the client
// this table should not continue to grow in size for repeated Posts.
const TaxonomyKind = "Taxonomy"

type Taxonomy struct {
	Key        *datastore.Key `datastore:"-"`
	BoardKey   string         // board this taxonomy is a member of
	Namespace  string         // parent namespace, empty string for top-level namespaces
	Childspace string         // a single child namespace of Namespace field
}

func storeMetric(c appengine.Context, boardKeyString string, postKey string, data PostBody, ancestor *Metric) (string, error) {
	metric := Metric{}
	metric.Namespace = data.Namespace
	if ancestor != nil {
		metric.Namespace = fmt.Sprintf("%s.%s", ancestor.Namespace, metric.Namespace)
	}
	metadata := ""
	if len(data.Meta) > 0 {
		if meta, err := json.Marshal(data.Meta); err == nil {
			metadata = string(meta)
		}
	}
	metric.Meta = metadata
	metric.Start = data.Start
	metric.End = data.End

	// compose an idempotent key for this data (allows Post data to be digested repeatedly)
	keyID := fmt.Sprintf("%s:%s:%v:%v", postKey, metric.Namespace, data.Start, data.End)
	if ancestor == nil {
		boardKey, _ := datastore.DecodeKey(boardKeyString)
		metric.Key = datastore.NewKey(c, MetricKind, keyID, 0, boardKey)
	} else {
		metric.Key = datastore.NewKey(c, MetricKind, keyID, 0, ancestor.Key)
	}
	_, err := datastore.Put(c, metric.Key, &metric)
	if err != nil {
		c.Errorf("Error on Metric Put:%v", err)
		return "", err
	}
	// recursive storage for child objects to the same entity table
	for _, child := range data.Children {
		if childNamespace, err := storeMetric(c, boardKeyString, postKey, child, &metric); err == nil {
			storeTaxonomy(c, boardKeyString, metric.Namespace, childNamespace)
		} else {
			c.Errorf("Error storing Taxonomy:%v", err)
		}
	}
	return metric.Namespace, nil
}

func storeTaxonomy(c appengine.Context, boardKey string, parentNamespace string, childNamespace string) {
	taxonomy := Taxonomy{}
	keyID := fmt.Sprintf("%s:%s:%s", boardKey, parentNamespace, childNamespace)
	taxonomy.Key = datastore.NewKey(c, TaxonomyKind, keyID, 0, nil)
	taxonomy.BoardKey = boardKey
	taxonomy.Namespace = parentNamespace
	taxonomy.Childspace = childNamespace
	if _, err := datastore.Put(c, taxonomy.Key, &taxonomy); err != nil {
		c.Errorf("Error on Taxonomy Put:%v", err)
	}
}

var digestPost = delay.Func("key", func(c appengine.Context, postKeyString string) {
	post := Post{}
	postKey, _ := datastore.DecodeKey(postKeyString)
	if err := datastore.Get(c, postKey, &post); err != nil {
		return
	}
	body := PostBody{}
	if err := json.Unmarshal([]byte(post.Body), &body); err != nil {
		c.Errorf("Error in Unmarshal:%v", err)
		return
	}

	// enter the recursive storage routine
	boardKeyString := postKey.Parent().Encode()
	if namespace, err := storeMetric(c, boardKeyString, postKeyString, body, nil); err == nil {
		storeTaxonomy(c, boardKeyString, "", namespace)
	} else {
		c.Errorf("Error storing Taxonomy:%v", err)
	}
})
