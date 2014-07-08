package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"github.com/gorilla/mux"
	"github.com/matthewbelisle-wf/jsonproperty"
	"io/ioutil"
	"net/http"
	"time"
)

// A Metric is a namespaced measurement. A metric's parent (for Ancestor queries)
// is the metric that contained it in the origional Post. Therefore, a Metric
// with the namespace 'X.Y.Z' will have an Ancestor Metric with a namespace
// of 'X.Y'.
const MetricKind = "Metric"

type Metric struct {
	Namespace string                    `datastore:"namespace" json:"namespace"`
	Meta      jsonproperty.JsonProperty `datastore:"-" json:"meta" jsonproperty:"meta"`
	Start     time.Time                 `datastore:"start" json:"start"`
	End       time.Time                 `datastore:"end" json:"end"`
	Depth     int64                     `datastore:"depth" json:"-"`
	Children  Metrics                   `datastore:"-" json:"children"`
}

type Metrics []*Metric

// Load() and Save() handle the meta property.

func (m *Metric) Load(c <-chan datastore.Property) error {
	c, err := jsonproperty.LoadJsonProperties(m, c)
	if err != nil {
		return err
	}
	return datastore.LoadStruct(m, c)
}

func (m *Metric) Save(c chan<- datastore.Property) error {
	c, err := jsonproperty.SaveJsonProperties(m, c)
	if err != nil {
		return err
	}
	return datastore.SaveStruct(m, c)
}

// Get() and Put() recursively gets/puts metrics into the datastore
func (m *Metric) GetChildren(c appengine.Context, key *datastore.Key, depth int64) (<-chan Metrics, <-chan error) {
	c1 := make(chan Metrics, 1)
	c2 := make(chan error, 1)
	q := datastore.NewQuery(MetricKind).Ancestor(key)
	if depth >= 0 {
		q.Filter("depth <=", m.Depth + depth)
	}

	hierarchy := map[string]*Metric{} // {keyString: *metric}
	for t := q.Run(c);; {
		child := Metric{}
		childKey, err := t.Next(&child)
		if err == datastore.Done {
			break
		} else if err != nil {
			c1 <- nil
			c2 <- err
			return c1, c2
		}
		hierarchy[childKey.Encode()] = &child
	}
	for keyString, child := range hierarchy {
		childKey, _ := datastore.DecodeKey(keyString)
		parentKey := childKey.Parent()
		if parent, ok := hierarchy[parentKey.Encode()]; ok {
			parent.Children = append(parent.Children, child)
		}
	}

	c1 <- hierarchy[key.Encode()].Children // Top level metric
	c2 <- nil
	return c1, c2
}

func (m *Metric) Put(c appengine.Context, key *datastore.Key) error {
	if _, err := datastore.Put(c, key, m); err != nil {
		c.Errorf("Failed metric.Put()\nCould not do datastore.Put(): %s, %s", key.Encode(), err)
		return err
	}
	for i, child := range m.Children {
		childKey := datastore.NewKey(c, MetricKind, child.Namespace, int64(i), key)
		child.Namespace = m.Namespace + "." + child.Namespace
		child.Depth = m.Depth + 1
		if err := child.Put(c, childKey); err != nil {
			c.Errorf("Failed metric.Put()\nCould not do child.Put(): %s, %s", childKey.Encode(), err)
			return err
		}
	}
	return nil
}

// HTTP handlers

func getMetric(w http.ResponseWriter, r *http.Request) {
	keyString := mux.Vars(r)["metric"]
	key, err := datastore.DecodeKey(keyString)
	if err != nil || key.Kind() != MetricKind {
		http.Error(w, "Invalid metrics key: %s"+keyString, http.StatusBadRequest)
		return
	}
	c := appengine.NewContext(r)
	metric := Metric{}
	if err = datastore.Get(c, key, &metric); err != nil {
		http.Error(w, "Metric not found: "+keyString, http.StatusNotFound)
		return
	}

	mChan, eChan := metric.GetChildren(c, key, -1)
	if err := <-eChan; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	metric.Children = <-mChan
	WriteJson(metric, w)
}

func getMetrics(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	vars := mux.Vars(r)
	encodedKey := vars["board"]
	boardKey, err := datastore.DecodeKey(encodedKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	board := Board{}
	if err := datastore.Get(c, boardKey, &board); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Assembles query
	q := datastore.NewQuery(MetricKind).
		Filter("namespace =", vars["namespace"])

	c.Infof("Filtering on namespaces: %v", string(vars["namespace"]))

	// Oldest point in request timeline
	if param := r.FormValue("start"); param != "" {
		if start, err := time.Parse(time.RFC3339, param); err != nil {
			http.Error(w, "Invalid start param: "+param, http.StatusBadRequest)
			return
		} else {
			q.Filter("start >=", start)
		}
	}

	// Newest point in request timeline
	if param := r.FormValue("end"); param != "" {
		if end, err := time.Parse(time.RFC3339, param); err != nil {
			http.Error(w, "Invalid end param: "+param, http.StatusBadRequest)
			return
		} else {
			q.Filter("end <=", end)
		}
	}

	q.Ancestor(boardKey) // Must be last?

	// Assembles metrics
	metrics := Metrics{}
	childrenChans := []<-chan Metrics{}
	errChans := []<-chan error{}
	t := q.Run(c)
	for {
		metric := Metric{}
		key, err := t.Next(&metric)
		if err == datastore.Done {
			break
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		metrics = append(metrics, &metric)
		cChan, eChan := metric.GetChildren(c, key, 1)
		childrenChans = append(childrenChans, cChan)
		errChans = append(errChans, eChan)
	}

	for i, metric := range metrics {
		if err := <-errChans[i]; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		metric.Children = <-childrenChans[i]
	}

	Json{
		"board": boardKey,
		"metrics": metrics,
	}.Write(w)
}

// TODO memcache the board entity and validate boardKey against it
func postMetric(w http.ResponseWriter, r *http.Request) {
	// Checks that the key is valid
	encodedBoardKey := mux.Vars(r)["board"]
	boardKey, err := datastore.DecodeKey(encodedBoardKey)
	if err != nil || boardKey.Kind() != BoardKind {
		http.Error(w, "Invalid Board key: "+encodedBoardKey, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	post := Post{
		Body:      string(body),
		Timestamp: time.Now(),
	}
	c := appengine.NewContext(r)
	postKey, err := datastore.Put(c, datastore.NewIncompleteKey(c, PostKind, boardKey), &post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	digestPost.Call(c, postKey)
}
