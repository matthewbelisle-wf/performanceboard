package performanceboard

import (
	"appengine"
	"appengine/aetest"
	"appengine/datastore"
	"encoding/json"
	// "io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type BoardResult struct {
	Api   string `json:"api"`
	Board string `json:"board"`
}

func createTestBoard(c appengine.Context) (*httptest.ResponseRecorder, error) {

	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "http://test.test/api/", nil)
	if err != nil {
		return nil, err
	}
	createBoard(c, w, req)

	// this block has a side effect of pushing data to disk
	var result BoardResult
	json.Unmarshal([]byte(w.Body.String()), &result)
	fetchBoard(c, result.Board)

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
	c, _ := aetest.NewContext(nil)
	w, _ := createTestBoard(c)

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
	c, _ := aetest.NewContext(nil)
	createTestBoard(c)
	createTestBoard(c)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://test.test/api/", nil)
	if err != nil {
		t.Fatal(err)
	}

	listBoards(c, w, req)
	t.Log(w.Body.String())

	var results []BoardResult
	err = json.Unmarshal([]byte(w.Body.String()), &results)
	log.Println(results)
	if len(results) != 2 {
		t.Fatal("expected 2 board result")
	}
}

func TestClearBoard(t *testing.T) {

}
