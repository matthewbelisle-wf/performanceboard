package performanceboard

import (
	"appengine"
	"appengine/aetest"
	"appengine/datastore"
	"encoding/json"
	"log"
	"net/http/httptest"
	"strings"
	"testing"
)

type BoardResult struct {
	Api   string `json:"api"`
	Board string `json:"board"`
}

func createTestBoard(inst aetest.Instance) (*httptest.ResponseRecorder, error) {
	w := httptest.NewRecorder()

	req, err := inst.NewRequest("POST", "http://test.test/api/", nil)
	if err != nil {
		return nil, err
	}

	context := appengine.NewContext(req)

	createBoard(context, w, req)

	// this block has a side effect of pushing data to disk
	var result BoardResult
	json.Unmarshal([]byte(w.Body.String()), &result)
	fetchBoard(context, result.Board)

	return w, nil
}

func fetchBoard(c appengine.Context, boardKey string) (*Board, error) {
	key, _ := datastore.DecodeKey(boardKey)
	board := Board{Key: key}
	if err := datastore.Get(c, key, &board); err != nil {
		return nil, err
	}
	return &board, nil
}

func TestCreateBoard(t *testing.T) {
	inst, _ := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	defer inst.Close()

	w, _ := createTestBoard(inst)

	if w.Code != 200 {
		t.Fatal("Expected code 200, recieved", w.Code)
	}

	var result BoardResult
	json.Unmarshal([]byte(w.Body.String()), &result)

	boardUrl := result.Api
	boardKey := result.Board

	if len(boardUrl) <= len(boardKey) || !strings.HasSuffix(boardUrl, boardKey) {
		t.Fatal("board url has unexpected format:", boardUrl)
	}
}

func TestListBoard(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	defer inst.Close()

	// setup preconditions
	createTestBoard(inst)
	createTestBoard(inst)

	// define call under test
	req, err := inst.NewRequest("GET", "http://test.test/api/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// make call
	c := appengine.NewContext(req)
	w := httptest.NewRecorder()
	listBoards(c, w, req)

	// define expected result structure
	type BoardResultResponse struct {
		Results []BoardResult `json:"results"`
	}
	var results BoardResultResponse
	err = json.Unmarshal([]byte(w.Body.String()), &results)

	// make assertions
	if len(results.Results) != 2 {
		t.Fatal("expected 2 board result, got ", len(results.Results))
	}
}

func TestClearBoard(t *testing.T) {
	t.Fatal()
}
