package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"appengine/delay"
	"encoding/json"
	"fmt"
	"time"
)

const BATCH_BLOCK_SIZE = 100

// data package parsing structure, not stored to disk.
type PostBody struct {
	Namespace string                 `json:"namespace"`
	Start     time.Time              `json:"start"`
	End       time.Time              `json:"end"`
	Meta      map[string]interface{} `json:"meta"`
	Children  []PostBody             `json:"children"`
}

// An AggregateMetric is a computed min, mean and max over a time interval
// second is the smallest block of time; minute, hour and day measure are
// derived from the smaller bins
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

const AggregateQueueKind = "AggregateQueue"

type AggregateQueue struct {
	PostKey   string
	Timestamp time.Time
}

func storeTaxonomy(context appengine.Context, boardKey string, parentNamespace string, childNamespace string) {
	taxonomy := Taxonomy{}
	keyID := fmt.Sprintf("%s:%s:%s", boardKey, parentNamespace, childNamespace)
	taxonomy.Key = datastore.NewKey(context, TaxonomyKind, keyID, 0, nil)
	taxonomy.BoardKey = boardKey
	taxonomy.Namespace = parentNamespace
	taxonomy.Childspace = childNamespace
	if _, err := datastore.Put(context, taxonomy.Key, &taxonomy); err != nil {
		context.Errorf("Error on Taxonomy Put:%v", err)
	}
}

// decompose a raw Post event into its hierarchical namespaces and write the Metrics to disk
func storeMetric(context appengine.Context, boardKeyString string, postKey string, data PostBody, ancestor *Metric, putBatch *PutBatch) (string, error) {
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
		metric.Key = datastore.NewKey(context, MetricKind, keyID, 0, boardKey)
	} else {
		metric.Key = datastore.NewKey(context, MetricKind, keyID, 0, ancestor.Key)
	}

    // store the metric in memory for a batch write
	putBatch.AddMetric(&metric)

	// recursive storage for child objects to the same entity table
	for _, child := range data.Children {
		if childNamespace, err := storeMetric(context, boardKeyString, postKey, child, &metric, putBatch); err == nil {
			// TODO batch taxonomy
			storeTaxonomy(context, boardKeyString, metric.Namespace, childNamespace)
		} else {
			context.Errorf("Error storing Taxonomy:%v", err)
		}
	}

	// TODO optimize how often the aggregate chain is called to once per namespace per post
	putBatch.AddAggBatchItem(metric.Start, boardKeyString, metric.Namespace)
	// aggregateSecond(context, metric.Start, boardKeyString, metric.Namespace)

	return metric.Namespace, nil
}

func aggregateSecond(context appengine.Context, t time.Time, boardKeyString string, namespace string, aggMetricsBatch *AggregateMetricBatch) {
	// context.Infof("aggregateSecond:%v", t)
	// Read the metrics table for a one-second interval
	boardKey, err := datastore.DecodeKey(boardKeyString)
	if err != nil {
		context.Errorf("aggregateSecond failed to decode %s: %v", boardKeyString, err)
		return
	}
	// trim fractional second to bin aggregate computation
	truncTime := t.Truncate(1 * time.Second)
	metrics, _, err := readMetrics(context, boardKey, namespace, truncTime, 1*time.Second, 0, -1, "")
	count := len(metrics)
	if count == 0 {
		return
	}

	// compute min, max, mean, and sample-count
	duration := metrics[0].End.Sub(metrics[0].Start)
	min := duration
	max := duration
	sum := time.Duration(0)
	for _, metric := range metrics {
		duration = metric.End.Sub(metric.Start)
		if duration < min {
			min = duration
		}
		if duration > max {
			max = duration
		}
		sum = sum + duration
	}

	// store values to AggregateMetric entity
	keyID := fmt.Sprintf("%s:%s:%v:%s", boardKeyString, namespace, truncTime, "second")
	key := datastore.NewKey(context, AggregateMetricKind, keyID, 0, boardKey)
	aggMetric := AggregateMetric{
		Key:       key,
		BoardKey:  boardKeyString,
		Namespace: namespace,
		StartTime: truncTime,
		BinType:   "second",
		Min:       float64(min) / float64(1000000.0),
		Max:       float64(max) / float64(1000000.0),
		Mean:      float64(sum) / float64(count) / float64(1000000.0), //convert to fractional milliseconds
		Sum:       int64(sum),
		Count:     int64(count),
	}

    // store the metric in memory for a batch write
    aggMetricsBatch.Add(&aggMetric)

	aggregateMore(context, t, boardKeyString, namespace, "minute", aggMetricsBatch)
}

func aggregateMore(context appengine.Context, t time.Time, boardKeyString string, namespace string, binType string, aggMetricsBatch *AggregateMetricBatch) {
	// context.Infof("aggregateMore:%s,%v", binType, t)
	// Read the AggregateMetric table for a one-minute interval
	boardKey, _ := datastore.DecodeKey(boardKeyString)
	var aggregateBinType string
	aggregateDuration := time.Minute
	switch binType {
	case "minute":
		aggregateBinType = "second"
	case "hour":
		aggregateBinType = "minute"
		aggregateDuration = time.Hour
	case "day":
		aggregateBinType = "hour"
		aggregateDuration = time.Duration(24) * time.Hour
	default:
		context.Errorf("Invalid aggregate bin type: %s, for board:%s", binType, boardKeyString)
		return
	}
	truncTime := t.Truncate(aggregateDuration)

	// readAggregates reads time backwards, so we jump time forward to read our duration
	aggregateMetrics, _, err := readAggregates(
		context, boardKey, namespace, aggregateBinType,
		truncTime.Add(aggregateDuration), aggregateDuration, -1, "")

	if err != nil {
		context.Errorf("aggregateMore failed to readAggregates on board: %s, %v", boardKeyString, err)
		return
	}
	if len(aggregateMetrics) == 0 {
		return
	}

	// compute min, max, mean, and sample-count
	min := aggregateMetrics[0].Min
	max := aggregateMetrics[0].Max
	sum := int64(0)
	count := int64(0)
	for _, aggregateMetric := range aggregateMetrics {
		if aggregateMetric.Min < min {
			min = aggregateMetric.Min
		}
		if aggregateMetric.Max > max {
			max = aggregateMetric.Max
		}
		sum = sum + aggregateMetric.Sum
		count = count + aggregateMetric.Count
	}

	// store values to AggregateMetric entity
	keyID := fmt.Sprintf("%s:%s:%v:%s", boardKeyString, namespace, truncTime, binType)
	key := datastore.NewKey(context, AggregateMetricKind, keyID, 0, boardKey)
	aggMetric := AggregateMetric{
		Key:       key,
		BoardKey:  boardKeyString,
		Namespace: namespace,
		StartTime: truncTime,
		BinType:   binType,
		Min:       float64(min),
		Max:       float64(max),
		Mean:      float64(sum) / float64(count) / float64(1000000.0), //convert to fractional milliseconds
		Sum:       int64(sum),
		Count:     int64(count),
	}

    // store the metric in memory for a batch write
    aggMetricsBatch.Add(&aggMetric)

	switch binType {
	case "minute":
		aggregateMore(context, t, boardKeyString, namespace, "hour", aggMetricsBatch)
	case "hour":
		aggregateMore(context, t, boardKeyString, namespace, "day", aggMetricsBatch)
	}

}

// entry point to decompose a post into
func digestPost(context appengine.Context, postKeyString string) *AggregationBatch {
	post := Post{}
	postKey, _ := datastore.DecodeKey(postKeyString)
	if err := datastore.Get(context, postKey, &post); err != nil {
		return nil
	}
	body := PostBody{}
	if err := json.Unmarshal([]byte(post.Body), &body); err != nil {
		context.Errorf("Error in Unmarshal:%v", err)
		return nil
	}

	// enter the recursive storage routine
	boardKeyString := postKey.Parent().Encode()

	putBatch := NewPutBatch()
	if namespace, err := storeMetric(context, boardKeyString, postKeyString, body, nil, putBatch); err == nil {
		storeTaxonomy(context, boardKeyString, "", namespace)

        // write this metric and its children to datastore
		keys, metrics := putBatch.GetMetrics()
        for i := 0; i < len(keys); i = i + BATCH_BLOCK_SIZE {
            stopIdx := i + BATCH_BLOCK_SIZE
            if stopIdx > len(keys) {
                stopIdx = len(keys)
            }

    		if _, err := datastore.PutMulti(context, keys[i:stopIdx], metrics[i:stopIdx]); err != nil {
    			context.Errorf("Error on Metric PutMulti:%v", err)
    		}
        }

	} else {
		context.Errorf("Error storing Taxonomy:%v", err)
	}
    return &putBatch.Aggregations
}

// This is the entry point into the deferred context of input digestion
// it fetchs all the keys to be processed, iterates, and deletes keys from the table
var digestPostQueue = delay.Func("key", func(context appengine.Context, postKeyString string) error {
	time.Sleep(100 * time.Millisecond)
	context.Infof("Running delayed process on AggregateQueue")
	queuedItems := []AggregateQueue{}
    aggBatch := AggregationBatch{}
	if keys, err := datastore.NewQuery(AggregateQueueKind).GetAll(context, &queuedItems); err == nil {
		for i, queuedItem := range queuedItems {
			context.Infof("processing post #%d, %s", i, queuedItem.PostKey)
			aggBatch.Merge(digestPost(context, queuedItem.PostKey))
			datastore.Delete(context, keys[i])
			context.Infof("finished wth post #%d, %s", i, queuedItem.PostKey)
		}
	}

    // process all the aggregations into entitys
    context.Infof("aggregation task count: %d", len(aggBatch))
    aggMetricsBatch := AggregateMetricBatch{}
    for _, item := range aggBatch {
        aggregateSecond(context, item.TimeStamp, item.BoardKey, item.Namespace, &aggMetricsBatch)
    }

    keys, metrics := aggMetricsBatch.GetMetrics()
    context.Infof("aggregation record count: %d", len(keys))
    for i := 0; i < len(keys); i = i + BATCH_BLOCK_SIZE {
        stopIdx := i + BATCH_BLOCK_SIZE
        if stopIdx > len(keys) {
            stopIdx = len(keys)
        }

        context.Infof("before agg batch put")
        if _, err := datastore.PutMulti(context, keys[i:stopIdx], metrics[i:stopIdx]); err != nil {
            context.Errorf("Error on Metric PutMulti:%v", err)
        }
        context.Infof("after agg batch put")
    }

	return nil // implies everything went well
})
