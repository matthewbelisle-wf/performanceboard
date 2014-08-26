package performanceboard

import (
    "appengine/datastore"
	"fmt"
    "time"
)

type AggBatchItem struct {
	TimeStamp      time.Time
	BoardKey  string
	Namespace string
}

type AggregationBatch map[string]AggBatchItem

func (ab *AggregationBatch) Contains(key string) bool {
    _, found := (*ab)[key]
    return found
}

func (ab *AggregationBatch) Add(t time.Time, boardKey, namespace string) {
    key := fmt.Sprintf("%s.%s.%s", t.Truncate(time.Second).String(), boardKey, namespace)
    if !ab.Contains(key) {
        (*ab)[key] = AggBatchItem{
            TimeStamp:      t,
            BoardKey:  boardKey,
            Namespace: namespace,
        }
    }
}

func (ab *AggregationBatch) Merge(other_ab *AggregationBatch) {
    if other_ab == nil {
        return
    }

    for _, item := range *other_ab {
        (*ab).Add(item.TimeStamp, item.BoardKey, item.Namespace)
    }
}

type AggregateMetricBatch []*AggregateMetric 

func (amb *AggregateMetricBatch) Add(aggMetric *AggregateMetric) {
    (*amb) = append(*amb, aggMetric)
}

func (amb *AggregateMetricBatch) GetMetrics() ([]*datastore.Key, []*AggregateMetric) {
    keys := []*datastore.Key{}
    for _, metric := range (*amb) {
        keys = append(keys, metric.Key)
    }
    return keys, ([]*AggregateMetric)(*amb)
}

type PutBatch struct {
    Metrics      []*Metric
    Aggregations AggregationBatch
}

func NewPutBatch() *PutBatch {
    putBatch := PutBatch{}
    putBatch.Aggregations = make(AggregationBatch)
    return &putBatch
}

func (pb *PutBatch) AddMetric(metric *Metric) {
	(*pb).Metrics = append((*pb).Metrics, metric)
}

func (pb *PutBatch) GetMetrics() ([]*datastore.Key, []*Metric) {
	keys := []*datastore.Key{}
	for _, metric := range (*pb).Metrics {
		keys = append(keys, metric.Key)
	}
	return keys, (*pb).Metrics
}

func (pb *PutBatch) AddAggBatchItem(t time.Time, boardKey, namespace string) {
	(*pb).Aggregations.Add(t, boardKey, namespace)
}
