package tests

import (
	"context"
	"testing"

	"github.com/ivanmartinez/strki"
	"github.com/ivanmartinez/strki/database"
)

func initializedDB() (db *MockDB) {
	return NewDB(map[string]database.Record{
		"rcrd1": database.Record{
			DbID: "rcrd1",
			FieldValues: map[string]string{
				"name":   "Karen Summers",
				"age":    "41",
				"gender": "F",
			},
		},
		"rcrd2": database.Record{
			DbID: "rcrd2",
			FieldValues: map[string]string{
				"name":   "Mark Smith",
				"age":    "18",
				"gender": "M",
			},
		},
		"rcrd3": database.Record{
			DbID: "rcrd3",
			FieldValues: map[string]string{
				"name":   "Alfred Hopkins",
				"age":    "60",
				"gender": "N",
			},
		},
	},
	)
}

func TestList(t *testing.T) {
	db := initializedDB()
	tData := strki.List(context.TODO(), db, "", nil).(strki.TemplateList)

	if record1, found := findRecord(tData.Records, "rcrd1"); found {
		checkRecordValue(t, record1, "name", "Karen Summers")
	} else {
		t.Error("Couldn't find \"rcrd1\"")
	}
	if record2, found := findRecord(tData.Records, "rcrd2"); found {
		checkRecordValue(t, record2, "age", "18")
	} else {
		t.Error("Couldn't find \"rcrd2\"")
	}
	if record3, found := findRecord(tData.Records, "rcrd3"); found {
		checkRecordValue(t, record3, "gender", "N")
	} else {
		t.Error("Couldn't find \"rcrd3\"")
	}
}

func findRecord(records []database.Record, DbID string) (database.Record, bool) {
	for _, record := range records {
		if record.DbID == DbID {
			return record, true
		}
	}
	return database.Record{}, false
}

func checkRecordValue(t *testing.T, record database.Record, field, value string) {
	if fieldValue, found := record.FieldValues[field]; !found {
		t.Errorf("Field \"%v\" not found", field)
	} else if fieldValue != value {
		t.Errorf("Field \"%v\" should be \"%v\" but is \"%v\"", field, value, fieldValue)
	}
}
