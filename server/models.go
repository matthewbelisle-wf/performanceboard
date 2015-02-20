package performanceboard

import (
	"appengine/datastore"
	"time"
)

// Board is the root structure for data. All data is posted against a board
const BoardKind = "Board"

type Board struct {
	Key    *datastore.Key `datastore:"-"`
	Name   string
	UserID string
}

// data package parsing structure, not stored to disk.
// it represents a metric
type PostBody struct {
	Namespace string                 `json:"namespace"`
	Start     time.Time              `json:"start"`
	End       time.Time              `json:"end"`
	Meta      map[string]interface{} `json:"meta"`
	Children  []PostBody             `json:"children"`
}

// A Post entity preserves a PostBody to datastore.
const PostKind = "Post"

type Post struct {
	Key       *datastore.Key `datastore:"-"`
	BoardKey  string         `datastore:",index"`
	Body      string         `datastore:",noindex"`
	Timestamp time.Time      `datastore:",noindex"`
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
	Children  []Metric       `datastore:"-"`
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

// AggregateMetric is a computed entity that captures the value of multiple
// Metric entities over an interval of time.
const AggregateMetricKind = "AggregateMetric"

type AggregateMetric struct {
	Key       *datastore.Key `datastore:"-"`
	BoardKey  string
	Namespace string
	StartTime time.Time
	BinType   string // second, minute, hour, day
	Min       float64
	Max       float64
	Mean      float64
	Sum       int64
	Count     int64
}
