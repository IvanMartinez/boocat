package server

import (
	"context"

	"github.com/ivanmartinez/boocat/log"
	"github.com/ivanmartinez/boocat/server/database"
	"github.com/ivanmartinez/boocat/server/formats"
)

var db database.DB

// Initialize initializes the server configuration without starting it. This is used in testing.
func Initialize(database database.DB) {
	db = database
}

// GetRecord returns a record of a format (author, book...)
func GetRecord(ctx context.Context, formatName string, record map[string]string) map[string]string {
	format := formats.Formats[formatName] // TODO: Check that the format exists
	// If record contains all the fields (from previous template), don't query the database
	record = fillFromDatabase(ctx, record, format)
	return record
}

// ListRecords returns a slice of all records of a format (authors, books...)
func ListRecords(ctx context.Context, formatName string) []map[string]string {
	records, err := db.GetAllRecords(ctx, formatName)
	if err != nil {
		log.Error.Printf("getting records from database: %v\n", err)
		return nil
	}
	return records
}

// SearchRecords returns a slice of the records of a format (authors, books...) whose search fields contains the value
func SearchRecords(ctx context.Context, formatName string, search string) []map[string]string {
	records, err := db.SearchRecord(ctx, formatName, search)
	if err != nil {
		log.Error.Printf("searching records in database: %v\n", err)
		return nil
	}
	return records
}

// AddRecord adds a record of a format (author, book...)
func AddRecord(ctx context.Context, formatName string, record map[string]string) map[string]string {
	format := formats.Formats[formatName] // TODO: Check that the format exists
	tplData := make(map[string]string)
	failed := format.Validate(ctx, record)
	tplData = add(tplData, failed)
	if len(failed) == 0 {
		id, err := db.AddRecord(ctx, format.Name, record)
		if err != nil {
			log.Error.Printf("adding record to database: %v\n", err)
		} else {
			tplData["id"] = id
			// Underscore value because empty string is empty pipeline in the template
			tplData["_success"] = "_"
		}
	}
	tplData = add(tplData, record)
	return tplData
}

// UpdateRecord updates a record of a format (author, book...)
func UpdateRecord(ctx context.Context, formatName string, record map[string]string) map[string]string {
	format := formats.Formats[formatName] // TODO: Check that the format exists
	tplData := make(map[string]string)
	failed := format.Validate(ctx, record)
	tplData = add(tplData, failed)
	if len(failed) == 0 {
		record = fillFromDatabase(ctx, record, format)
		if err := db.UpdateRecord(ctx, format.Name, record); err != nil {
			log.Error.Printf("updating record in database: %v\n", err)
		} else {
			// Underscore value because empty string is empty pipeline in the template
			tplData["_success"] = "_"
		}
	}
	tplData = add(tplData, record)
	return tplData
}

// fillFromDatabase If record is missing any field of the format, then get it from the database and return the filled
// record
func fillFromDatabase(ctx context.Context, record map[string]string, format formats.Format) map[string]string {
	if format.IncompleteRecord(record) {
		if dbRecord, err := db.GetRecord(ctx, format.Name, record["id"]); err == nil {
			record = format.Merge(record, dbRecord)
		} else {
			log.Error.Printf("getting record from database: %v\n", err)
		}
	}
	return record
}

// add adds the elements of sMap to pMap and returns the result. Keys that exist in both maps are left as they are in
// pMap
func add(pMap, sMap map[string]string) (tMap map[string]string) {
	for key, value := range sMap {
		if _, found := pMap[key]; !found {
			pMap[key] = value
		}
	}
	return pMap
}
