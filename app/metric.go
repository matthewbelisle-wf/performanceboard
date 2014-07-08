package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
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
	Key       *datastore.Key            `datastore:"-"`
	Context   appengine.Context         `datastore:"-"`
	Namespace string                    `datastore:"namespace"`
	Meta      jsonproperty.JsonProperty `datastore:"-" jsonproperty:"meta"`
	Start     time.Time                 `datastore:"start"`
	End       time.Time                 `datastore:"end"`
	Depth     int64                     `datastore:"depth"`
	Children  Metrics                   `datastore:"-"`
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

// Recursively loads children from the datastore up to a certain depth.  Must be fast!
func (m *Metric) LoadChildren(depth int64) <-chan error {
	errChan := make(chan error)
	go func() {
		q := datastore.NewQuery(MetricKind).Ancestor(m.Key)
		if depth >= 0 {
			q.Filter("depth <=", m.Depth+depth)
		}

		hierarchy := map[string]*Metric{} // {keyString: *metric}
		var err error
		for t := q.Run(m.Context); ; {
			child := Metric{Context: m.Context}
			child.Key, err = t.Next(&child)
			if err == datastore.Done {
				break
			} else if err != nil {
				errChan <- err
				return
			}
			hierarchy[child.Key.Encode()] = &child
		}
		for keyString, child := range hierarchy {
			childKey, _ := datastore.DecodeKey(keyString)
			parentKey := childKey.Parent()
			if parent, ok := hierarchy[parentKey.Encode()]; ok {
				parent.Children = append(parent.Children, child)
			}
		}

		m.Children = hierarchy[m.Key.Encode()].Children // Top level metric
		errChan <- nil
	}()
	return errChan
}

// Recursively puts children into the datastore.  Does not need to be fast.
func (m *Metric) Put() error {
	if _, err := datastore.Put(m.Context, m.Key, m); err != nil {
		m.Context.Errorf("Failed metric.Put()\nCould not do datastore.Put(): %s, %s", m.Key.Encode(), err)
		return err
	}
	for i, child := range m.Children {
		child.Key = datastore.NewKey(m.Context, MetricKind, child.Namespace, int64(i), m.Key)
		child.Context = m.Context
		// TODO: Move these to save?
		child.Namespace = m.Namespace + "." + child.Namespace
		child.Depth = m.Depth + 1
		if err := child.Put(); err != nil {
			return err
		}
	}
	return nil
}

func (m Metric) MarshalJSON() ([]byte, error) {
	api, _ := router.Get("metric").URL("metric", m.Key.Encode())
	return json.Marshal(Json{
		"key":      m.Key,
		"start":    m.Start,
		"end":      m.End,
		"meta":     m.Meta,
		"children": m.Children,
		"api":      AbsURL(api, m.Context.Request().(*http.Request)),
	})
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
	metric := Metric{Context: c, Key: key}
	if err = datastore.Get(c, key, &metric); err != nil {
		http.Error(w, "Metric not found: "+keyString, http.StatusNotFound)
		return
	}

	if err := <-metric.LoadChildren(-1); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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
	errChans := []<-chan error{}
	t := q.Run(c)
	for {
		metric := Metric{Context: c}
		metric.Key, err = t.Next(&metric)
		if err == datastore.Done {
			break
		} else if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		metrics = append(metrics, &metric)
		errChans = append(errChans, metric.LoadChildren(1))
	}

	for _, errChan := range errChans {
		if err := <-errChan; err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	Json{
		"board":   boardKey,
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
