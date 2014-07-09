package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"appengine/delay"
	"encoding/json"
	"fmt"
	"time"
)

// aggregateSecond(context, metric.Start, boardKeyString, metric.Namespace)

func readMetrics(context appengine.Context, boardKey *datastore.Key, namespace string,
	newestTime time.Time, duration time.Duration, depth int, limit int, cursor string) ([]Metric, string, error) {

	// Duration is used instead of oldestTime to help support unbound queries
	q := datastore.NewQuery(MetricKind).
		Filter("Namespace =", namespace).
		Filter("Start <=", newestTime).
		Order("-Start").
		Ancestor(boardKey)

	if duration > 0 {
		oldestTime := newestTime.Add(-duration)
		q = q.Filter("Start >=", oldestTime)
	}

	if len(cursor) > 0 {
		if c, err := datastore.DecodeCursor(cursor); err == nil {
			q = q.Start(c)
		} else {
			return []Metric{}, "", err
		}
	}

	var metrics []Metric
	iter := q.Run(context)
	for limit < 0 || len(metrics) < limit {
		var metric Metric
		if key, err := iter.Next(&metric); err == nil {
			metric.Key = key
		} else {
			break
		}
		if depth != 0 { // recursively fetch children
			if kids, err := readMetricChildren(context, boardKey.Encode(), &metric, (depth - 1)); err == nil {
				if len(kids) > 0 {
					metric.Children = kids
				}
			}
		}
		metrics = append(metrics, metric)
	}

	if len(metrics) == limit {
		if c, err := iter.Cursor(); err == nil {
			cursor = c.String()
		}
	} else {
		cursor = ""
	}

	return metrics, cursor, nil
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
	metrics, err := readMetrics(context, boardKey, namespace, truncTime, 1*time.Second)
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
		Start: truncTime,
		BinType:   "second",
		Min:       float64(min) / float64(1000000.0),
		Max:       float64(max) / float64(1000000.0),
		Mean:      float64(sum) / float64(count) / float64(1000000.0), //convert to fractional milliseconds
		Sum:       int64(sum),
		Count:     int64(count),
	}
	if _, err := datastore.Put(context, aggMetric.Key, &aggMetric); err != nil {
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
	aggregateMetrics, err := readAggregates(
		context, boardKey, namespace, aggregateBinType,
		truncTime.Add(aggregateDuration), aggregateDuration, 0, "")
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
		Start: truncTime,
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

// This is the entry point into the deferred context of input digestion
var digestPost = delay.Func("digestPost", func(c appengine.Context, postKey *datastore.Key) {
	post := Post{}
	if err := datastore.Get(c, postKey, &post); err != nil {
		c.Errorf("Failed digestPost(): %s\nCould not Get: %s", postKey, err)
		return
	}
	metric := Metric{Context: c}
	if err := json.Unmarshal([]byte(post.Body), &metric); err != nil {
		c.Errorf("Failed digestPost(): %s\nCould not Unmarshal: %s", postKey, err)
		return
	}
	metric.Key = datastore.NewKey(c, MetricKind, metric.Namespace, 0, postKey)
	if err := metric.Put(); err != nil {
		return
	}
})
