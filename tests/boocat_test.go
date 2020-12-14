package tests

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ivanmartinez/boocat/formats"
	"github.com/ivanmartinez/boocat/server"
)

// initiaziledDB returns a MockDB with data for testing
func initializedDB() (db *MockDB) {
	db = NewDB()
	db.AddRecord(context.TODO(), "author", map[string]string{
		"name":      "Haruki Murakami",
		"birthdate": "1949",
		"biography": "Japanese",
	})
	db.AddRecord(context.TODO(), "author", map[string]string{
		"name":      "George Orwell",
		"birthdate": "1903",
		"biography": "English",
	})
	db.AddRecord(context.TODO(), "author", map[string]string{
		"name":      "miguel de cervantes saavedra",
		"birthdate": "MDXLVII",
		"biography": "Spanish",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name":     "Norwegian Wood",
		"year":     "1987",
		"author":   "author1",
		"synopsis": "novel",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name":     "Kafka On The Shore",
		"year":     "2002",
		"author":   "author1",
		"synopsis": "novel",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name":     "Animal Farm",
		"year":     "1945",
		"author":   "author2",
		"synopsys": "fable",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name":   "Nineteen Eighty-Four",
		"year":   "1949",
		"author": "author2",
	})
	return db
}

func TestWrongMethod(t *testing.T) {
	req := httptest.NewRequest("PUT", "/", nil)
	res := handle(req)
	if res.StatusCode != 400 {
		t.Errorf("expected status code 400 but got %v", res.StatusCode)
	}
}

func TestNoFile(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	res := handle(req)
	if res.StatusCode != 404 {
		t.Errorf("expected status code 404 but got %v", res.StatusCode)
	}
}

func TestStaticFile(t *testing.T) {
	initialize()
	req := httptest.NewRequest("GET", "/hello", nil)
	res := handle(req)
	if res.StatusCode != 200 {
		t.Errorf("expected status code 200 but got %v", res.StatusCode)
	}
	buf := new(strings.Builder)
	if _, err := io.Copy(buf, res.Body); err != nil {
		t.Error(err)
	} else if strings.TrimSpace(buf.String()) != "hello" {
		t.Errorf("expected reading file with body \"hello\" but got \"%v\"", strings.TrimSpace(buf.String()))
	}
}

//@TODO: Check that a tamplate has the name of a format. I'm going to leave this for later because I will
// refactor this part.

func TestGetRecord(t *testing.T) {
	initialize()
	req := httptest.NewRequest("GET", "/author?id=author1", nil)
	res := handle(req)
	if res.StatusCode != 200 {
		t.Errorf("expected status code 200 but got %v", res.StatusCode)
	}
	resMap := decodeJSONBody(t, res.Body)
	checkMap(t, resMap, map[string]string{
		"name":      "Haruki Murakami",
		"birthdate": "1949",
		"biography": "Japanese"})
}

func decodeJSONBody(t *testing.T, reader io.ReadCloser) (m map[string]string) {
	t.Helper()
	defer reader.Close()
	// Decode the JSON
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(&m); err != nil {
		t.Errorf("couldn't decode JSON: %v", err)
	}
	return m
}

func checkMap(t *testing.T, subjectMap, checks map[string]string) {
	t.Helper()
	for key, checkValue := range checks {
		if value, found := subjectMap[key]; found {
			if value != checkValue {
				t.Errorf("field \"%v\" value \"%v\" should be \"%v\"", key, value, checkValue)
			}
		} else {
			t.Errorf("field \"%v\" not found", key)
		}
	}
}

func handle(req *http.Request) *http.Response {
	w := httptest.NewRecorder()
	server.Handle(w, req)
	return w.Result()
}

func initialize() {
	db := initializedDB()
	formats.Initialize(db)
	server.StartServer(context.Background(), "", "web", db)
}
