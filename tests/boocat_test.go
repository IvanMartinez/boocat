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

// initializedDB returns a MockDB with data for testing
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
		"name":     "Nineteen Eighty-Four",
		"year":     "1949",
		"author":   "author2",
		"synopsys": "dystopia",
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

func TestGetRecord(t *testing.T) {
	initialize()
	req := httptest.NewRequest("GET", "/author?id=author1", nil)
	res := handle(req)
	if res.StatusCode != 200 {
		t.Errorf("expected status code 200 but got %v", res.StatusCode)
	}
	fields := decodeToMap(t, res.Body)
	checkValues(t, fields, map[string]string{
		"id":        "author1",
		"name":      "Haruki Murakami",
		"birthdate": "1949",
		"biography": "Japanese"})
}

func TestGetRecords(t *testing.T) {
	initialize()
	req := httptest.NewRequest("GET", "/list/book", nil)
	res := handle(req)
	if res.StatusCode != 200 {
		t.Errorf("expected status code 200 but got %v", res.StatusCode)
	}
	resMaps := decodeToMaps(t, res.Body)
	if len(resMaps) != 4 {
		t.Errorf("expected 4 maps but got %v", len(resMaps))
	}
	findMapInSlice(t, resMaps, map[string]string{
		"name":     "Norwegian Wood",
		"year":     "1987",
		"author":   "author1",
		"synopsis": "novel"})
	findMapInSlice(t, resMaps, map[string]string{
		"name":     "Nineteen Eighty-Four",
		"year":     "1949",
		"author":   "author2",
		"synopsys": "dystopia"})
}

func TestSearchRecords(t *testing.T) {
	initialize()
	req := httptest.NewRequest("GET", "/list/author?_search=orwell", nil)
	res := handle(req)
	if res.StatusCode != 200 {
		t.Errorf("expected status code 200 but got %v", res.StatusCode)
	}
	resMaps := decodeToMaps(t, res.Body)
	if len(resMaps) != 1 {
		t.Errorf("expected 1 map but got %v", len(resMaps))
	}
	findMapInSlice(t, resMaps, map[string]string{
		"name":      "George Orwell",
		"birthdate": "1903",
		"biography": "English",
	})
}

func TestAddRecord(t *testing.T) {
	initialize()
	req := httptest.NewRequest("POST", "/book", strings.NewReader("name=The Wind-Up Bird Chronicle&year=1995&"+
		"author=author1&synopsis=novel"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := handle(req)
	if res.StatusCode != 200 {
		t.Errorf("expected status code 200 but got %v", res.StatusCode)
	}
	// Check the response to the request
	fields := decodeToMap(t, res.Body)
	checkValues(t, fields, map[string]string{
		"id":       "book5",
		"name":     "The Wind-Up Bird Chronicle",
		"year":     "1995",
		"author":   "author1",
		"synopsis": "novel",
		"_success": "_"})
	// Check getting the same record to make sure it's properly stored
	req = httptest.NewRequest("GET", "/book?id="+fields["id"], nil)
	res = handle(req)
	if res.StatusCode != 200 {
		t.Errorf("expected status code 200 but got %v", res.StatusCode)
	}
	fields = decodeToMap(t, res.Body)
	checkValues(t, fields, map[string]string{
		"id":       "book5",
		"name":     "The Wind-Up Bird Chronicle",
		"year":     "1995",
		"author":   "author1",
		"synopsis": "novel"})
}

func TestAddRecordValidationFail(t *testing.T) {
	initialize()
	req := httptest.NewRequest("POST", "/book", strings.NewReader("name=the wind-up bird chronicle&year=95&"+
		"author=author1&synopsis=novel"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := handle(req)
	if res.StatusCode != 200 {
		t.Errorf("expected status code 200 but got %v", res.StatusCode)
	}
	// Check the response to the request
	fields := decodeToMap(t, res.Body)
	checkValues(t, fields, map[string]string{
		"name":       "the wind-up bird chronicle",
		"year":       "95",
		"author":     "author1",
		"synopsis":   "novel",
		"_name_fail": "_",
		"_year_fail": "_",
	})
	// @TODO: Check that the record hasn't been added
}

func TestUpdateRecord(t *testing.T) {
	initialize()
	req := httptest.NewRequest("POST", "/author", strings.NewReader("id=author3&name=Miguel De Cervantes Saavedra&"+
		"birthdate=1547"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := handle(req)
	if res.StatusCode != 200 {
		t.Errorf("expected status code 200 but got %v", res.StatusCode)
	}
	// Check the response to the request
	fields := decodeToMap(t, res.Body)
	checkValues(t, fields, map[string]string{
		"id":        "author3",
		"name":      "Miguel De Cervantes Saavedra",
		"birthdate": "1547",
		"biography": "Spanish",
		"_success":  "_"})
	// Check getting the same record to make sure it's properly stored
	req = httptest.NewRequest("GET", "/author?id="+fields["id"], nil)
	res = handle(req)
	if res.StatusCode != 200 {
		t.Errorf("expected status code 200 but got %v", res.StatusCode)
	}
	fields = decodeToMap(t, res.Body)
	// Check the response to the request
	checkValues(t, fields, map[string]string{
		"id":        "author3",
		"name":      "Miguel De Cervantes Saavedra",
		"birthdate": "1547",
		"biography": "Spanish"})
}

func TestUpdateRecordValidationFail(t *testing.T) {
	initialize()
	req := httptest.NewRequest("POST", "/author", strings.NewReader("id=author2&name=george orwell&"+
		"birthdate=MCMIII"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := handle(req)
	if res.StatusCode != 200 {
		t.Errorf("expected status code 200 but got %v", res.StatusCode)
	}
	// Check the response to the request
	fields := decodeToMap(t, res.Body)
	checkValues(t, fields, map[string]string{
		"id":              "author2",
		"name":            "george orwell",
		"birthdate":       "MCMIII",
		"_name_fail":      "_",
		"_birthdate_fail": "_"})
	// Check getting the same record to make sure it hasn't changed
	req = httptest.NewRequest("GET", "/author?id="+fields["id"], nil)
	res = handle(req)
	if res.StatusCode != 200 {
		t.Errorf("expected status code 200 but got %v", res.StatusCode)
	}
	fields = decodeToMap(t, res.Body)
	checkValues(t, fields, map[string]string{
		"id":        "author2",
		"name":      "George Orwell",
		"birthdate": "1903",
		"biography": "English"})
}

func decodeToMap(t *testing.T, reader io.ReadCloser) (m map[string]string) {
	t.Helper()
	defer reader.Close()
	// Decode the JSON
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(&m); err != nil {
		t.Errorf("couldn't decode JSON: %v", err)
	}
	return m
}

func decodeToMaps(t *testing.T, reader io.ReadCloser) (s []map[string]string) {
	t.Helper()
	defer reader.Close()
	// Decode the JSON
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(&s); err != nil {
		t.Errorf("couldn't decode JSON: %v", err)
	}
	return s
}

// findMapInSlice checks that there is a map in subjectMaps that contains al the key-value pairs of checks
func findMapInSlice(t *testing.T, subjectMaps []map[string]string, checks map[string]string) {
	t.Helper()
	// For every map
	for _, subjectMap := range subjectMaps {
		match := true
		// For every check
		for key, checkValue := range checks {
			if value, found := subjectMap[key]; found {
				if value != checkValue {
					// subjectMap's value of key doesn't match checkValue
					match = false
					// We don't perform more checks on this map
					break
				}
			} else {
				// key doesn't exist in subjectMap
				match = false
				// We don't perform more checks on this map
				break
			}
		}
		if match {
			// All the checks matched, we found a matching map
			return
		}
	}
	t.Errorf("couldn't find a map that matches %v", checks)
}

// checkValues checks that the values of templateData matches the checks
func checkValues(t *testing.T, templateData map[string]string, checks map[string]string) {
	t.Helper()
	for key, checkValue := range checks {
		if value, found := templateData[key]; found {
			if value != checkValue {
				t.Errorf("field \"%v\" value \"%v\" should be \"%v\"", key, value, checkValue)
			}
		} else {
			t.Errorf("field \"%v\" not found", key)
		}
	}
	for key := range templateData {
		if _, found := checks[key]; !found {
			t.Errorf("field \"%v\" not expected", key)
		}
	}
}

func handle(req *http.Request) *http.Response {
	w := httptest.NewRecorder()
	server.Handle(w, req)
	return w.Result()
}

func initialize() {
	formats.Initialize()
	db := initializedDB()
	server.Initialize("", "web", db)
}
