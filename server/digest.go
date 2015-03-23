package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"appengine/delay"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

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
func storeMetric(context appengine.Context, boardKeyString string, postKey string, data PostBody, parent *Metric) (string, error) {
	metric := Metric{}
	metric.Namespace = data.Namespace
	if parent != nil {
		metric.Namespace = fmt.Sprintf("%s.%s", parent.Namespace, metric.Namespace)
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
	metric.BoardKey = boardKeyString
	if parent != nil {
		metric.ParentKey = parent.Key.Encode()
	}

	// compose an idempotent key for this data (allows Post data to be digested repeatedly)
	keyID := fmt.Sprintf("%s:%s:%v:%v", postKey, metric.Namespace, data.Start, data.End)
	metric.Key = datastore.NewKey(context, MetricKind, keyID, 0, nil)
	_, err := datastore.Put(context, metric.Key, &metric)
	if err != nil {
		log.Println("Error on Metric Put:", err)
		context.Errorf("Error on Metric Put:%v", err)
		return "", err
	}
	// recursive storage for child objects to the same entity table
	for _, child := range data.Children {
		if childNamespace, err := storeMetric(context, boardKeyString, postKey, child, &metric); err == nil {
			storeTaxonomy(context, boardKeyString, metric.Namespace, childNamespace)
		} else {
			context.Errorf("Error storing Taxonomy:%v", err)
		}
	}

	// TODO restore aggregation when it is qualified with tests
	// TODO optimize how often the aggregate chain is called to once per namespace per post
	// aggregateSecond(context, metric.Start, boardKeyString, metric.Namespace)

	return metric.Namespace, nil
}

func aggregateSecond(context appengine.Context, t time.Time, boardKeyString string, namespace string) {
	// Read the metrics table for a one-second interval
	boardKey, err := datastore.DecodeKey(boardKeyString)
	if err != nil {
		context.Errorf("aggregateSecond failed to decode %s: %v", boardKeyString, err)
		return
	}
	// trim fractional second to bin aggregate computation
	truncTime := t.Truncate(1 * time.Second)
	metrics, _, err := readMetrics(context, boardKeyString, namespace, truncTime, 1*time.Second, 0, -1, "")
	count := len(metrics)
	if count == 0 {
		log.Println("no data to aggregate")
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
	if _, err := datastore.Put(context, aggMetric.Key, &aggMetric); err != nil {
		log.Println("Error on Metric Put:%v", err)
		context.Errorf("Error on Metric Put:%v", err)
		return
	}

	aggregateMore(context, t, boardKeyString, namespace, "minute")
}

func aggregateMore(context appengine.Context, t time.Time, boardKeyString string, namespace string, binType string) {
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
	if _, err := datastore.Put(context, aggMetric.Key, &aggMetric); err != nil {
		context.Errorf("Error on Metric Put:%v", err)
		return
	}

	// TODO delay continued aggregation to let help let previous aggregation data settle
	switch binType {
	case "minute":
		aggregateMore(context, t, boardKeyString, namespace, "hour")
	case "hour":
		aggregateMore(context, t, boardKeyString, namespace, "day")
	}

}

func digestPost(c appengine.Context, postKeyString string) {
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
	if namespace, err := storeMetric(c, post.BoardKey, postKeyString, body, nil); err == nil {
		storeTaxonomy(c, post.BoardKey, "", namespace)
	} else {
		c.Errorf("Error storing Taxonomy:%v", err)
	}
}

// This is the entry point into the deferred context of input digestion
var delayedDigestPost = delay.Func("key", digestPost)
