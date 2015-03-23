package performanceboard

import (
	"appengine"
	"appengine/aetest"
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPostMetric(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	defer inst.Close()

	// setup preconditions
	response, err := createTestBoard(inst)
	if err != nil {
		t.Fatal(err)
	}

	var result BoardResult
	json.Unmarshal([]byte(response.Body.String()), &result)

	boardKeyString := result.Board
	t.Log("boardKeyString:", boardKeyString)

	body := `{
        "start":"1994-11-05T13:15:30Z",
        "end":"1994-11-05T13:15:31Z",
        "meta":{"meta":"data"},
        "namespace":"test"
        }`

	// define call under test
	req, err := inst.NewRequest("POST",
		"http://www.test.com/api/"+boardKeyString,
		bytes.NewBuffer([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}

	// make call
	w := httptest.NewRecorder()
	r := initRouter() // enables parsing route parameters
	r.ServeHTTP(w, req)


	postBody := PostBody{}
	if err := json.Unmarshal([]byte(body), &postBody); err != nil {
		t.Fatal(err)
	}

	// assert the call created an entity
	context := appengine.NewContext(req)
	metrics, cursor, err := readMetrics(
		context,
		boardKeyString,
		"test",        // namespace
		postBody.End,  // newestTime (rev chron order)
		3 * time.Second, // duration
		0,             // depth
		100,           // limit
		"")            // cursor
	if err != nil {
		t.Fatal(err)
	}
	if len(metrics) < 1 {
		t.Fatal("no metrics stored")
	}
	if metrics[0].Namespace != "test" {
		t.Fatal("error reading metric")
	}
	if cursor != "" {
		t.Fatal("non-empty cursor recieved")
	}
}
