package performanceboard

import (
	"appengine"
	"appengine/datastore"
	"github.com/gorilla/mux"
	"github.com/matthewbelisle-wf/jsonproperty"
	"io/ioutil"
	"net/http"
	// "strconv"
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

// Get() and Put() recursively gets/puts metrics into the datastore
func (m *Metric) Get(c appengine.Context, key *datastore.Key) error {
	q := datastore.NewQuery(MetricKind).Ancestor(key)
	hierarchy := map[string]*Metric{} // {keyString: *metric}
	for t := q.Run(c); ; {
		child := Metric{}
		childKey, err := t.Next(&child)
		if err == datastore.Done {
			break
		} else if err != nil {
			return err
		}
		hierarchy[childKey.Encode()] = &child
	}
	// Resetting hierarchy since ancestor filters limit results to the specified entity *and* its descendants.
	hierarchy[key.Encode()] = m

	for keyString, child := range hierarchy {
		childKey, _ := datastore.DecodeKey(keyString)
		parentKey := childKey.Parent()
		if parent, ok := hierarchy[parentKey.Encode()]; ok {
			parent.Children = append(parent.Children, *child)
		}
	}
	return nil
}

func (m *Metric) Put(c appengine.Context, key *datastore.Key) error {
	if _, err := datastore.Put(c, key, m); err != nil {
		c.Errorf("Failed metric.Put()\nCould not do datastore.Put(): %s, %s", key.Encode(), err)
		return err
	}
	for i, child := range m.Children {
		childKey := datastore.NewKey(c, MetricKind, child.Namespace, int64(i), key)
		child.Namespace = m.Namespace + "." + child.Namespace
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
	if err = metric.Get(c, key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	JsonResponse{
		"metric": metric,
	}.Write(w)
}

func getMetrics(w http.ResponseWriter, r *http.Request) {
	// c := appengine.NewContext(r)
	// vars := mux.Vars(r)
	// encodedKey := vars["board"]
	// boardKey, err := datastore.DecodeKey(encodedKey)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }
	// namespace := vars["namespace"]

	// // Evalutates oldest point in request timeline
	// var start time.Time
	// if startParam := r.FormValue("start"); startParam != "" {
	// 	if start, err = time.Parse(time.RFC3339, startParam); err != nil {
	// 		http.Error(w, "Invalid start param: "+startParam, http.StatusBadRequest)
	// 	}
	// }

	// // Evalutates newest point in request timeline
	// var end time.Time
	// if endParam := r.FormValue("end"); endParam != "" {
	// 	if end, err = time.Parse(time.RFC3339, endParam); err != nil {
	// 		http.Error(w, "Invalid end param: "+endParam, http.StatusBadRequest)
	// 	}
	// }

	// depth := int64(0)
	// if depthParam := r.FormValue("depth"); depthParam != "" {
	// 	if depth, err = strconv.ParseInt(depthParam, 10, 0); err != nil {
	// 		c.Errorf("Error parsing depth: %s", err)
	// 		http.Error(w, err.Error(), http.StatusBadRequest)
	// 		return
	// 	}
	// }
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
