package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"github.com/gorilla/mux"
	"github.com/matthewbelisle-wf/jsonproperty"
	"io/ioutil"
	"net/http"
	"strconv"
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
	Children  []Metric                  `datastore:"-" json:"children"`
}

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

// Get() and Put() are for recursive getting and putting

func (m *Metric) Get(c appengine.Context, key *datastore.Key, depth int) error {
	return nil
}

func (m *Metric) Put(c appengine.Context, key *datastore.Key) error {
	return nil
}

// HTTP handlers

func getMetrics(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	vars := mux.Vars(r)
	encodedKey := vars["board"]
	boardKey, err := datastore.DecodeKey(encodedKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	namespace := vars["namespace"]

	// Evalutates oldest point in request timeline
	var start time.Time
	if startParam := r.FormValue("start"); startParam != "" {
		if start, err = time.Parse(time.RFC3339, startParam); err != nil {
			http.Error(w, "Invalid start param: "+startParam, http.StatusBadRequest)
		}
	}

	// Evalutates newest point in request timeline
	var end time.Time
	if endParam := r.FormValue("end"); endParam != "" {
		if end, err = time.Parse(time.RFC3339, endParam); err != nil {
			http.Error(w, "Invalid end param: "+endParam, http.StatusBadRequest)
		}
	}

	depth := int64(0)
	if depthParam := r.FormValue("depth"); depthParam != "" {
		if depth, err = strconv.ParseInt(depthParam, 10, 0); err != nil {
			c.Errorf("Error parsing depth: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
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
	digestPost.Call(c, postKey.Encode())
}
