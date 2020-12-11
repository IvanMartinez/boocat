package boocat

import (
	"context"
	"log"

	"github.com/ivanmartinez/boocat/database"
	"github.com/ivanmartinez/boocat/formats"
)

// HTTPURL is the base URL for links and URLs in the HTML templates
// @TODO: Is There a better solution than this global variable?
var (
	HTTPURL string
	DB      database.DB
)

func Get(ctx context.Context, formatName string, params map[string]string) interface{} {
	if id, found := params["id"]; found {
		return getRecord(ctx, formatName, id)
	}
	return list(ctx, formatName)
}

func Update(ctx context.Context, formatName string, params map[string]string) interface{} {
	if _, found := params["id"]; found {
		return updateRecord(ctx, formatName, params)
	}
	return newRecord(ctx, formatName, params)
}

// newRecord adds a record of a format (author, book...)
func newRecord(ctx context.Context, formatName string, record map[string]string) interface{} {
	format, found := formats.Get(formatName)
	if !found {
		log.Printf("couldn't find format \"%v\"", formatName)
		return nil
	}
	failed := formats.Validate(ctx, format, record)
	if len(failed) == 0 {
		id, err := DB.AddRecord(ctx, formatName, record)
		if err != nil {
			log.Printf("error adding record to database: %v\n", err)
		} else {
			record["id"] = id
		}
	}
	return record
}

// updateRecord updates a record of a format (author, book...)
func updateRecord(ctx context.Context, formatName string, record map[string]string) interface{} {
	format, found := formats.Get(formatName)
	if !found {
		log.Printf("couldn't find format \"%v\"", formatName)
		return nil
	}
	failed := formats.Validate(ctx, format, record)
	if len(failed) == 0 {
		// If record doesn't have all the fields defined in the format, get the missing fields from the database
		if formats.IncompleteRecord(format, record) {
			if dbRecord, err := DB.GetRecord(ctx, formatName, record["id"]); err == nil {
				record = formats.Merge(format, record, dbRecord)
			} else {
				log.Printf("error getting database record: %v\n", err)
			}
		}
		if err := DB.UpdateRecord(ctx, formatName, record); err != nil {
			log.Printf("error updating record in database: %v\n", err)
		}
	}
	return record
}

// getRecord returns a record of a format (author, book...)
func getRecord(ctx context.Context, formatName, id string) map[string]string {
	record, err := DB.GetRecord(ctx, formatName, id)
	if err != nil {
		log.Printf("error getting database record: %v\n", err)
		//_, tplData := EditNew(ctx, format, id, nil)
		return nil
	}

	return record
}

// list returns a slice of all records of a format (authors, books...)
func list(ctx context.Context, format string) []map[string]string {
	records, err := DB.GetAllRecords(ctx, format)
	if err != nil {
		log.Printf("error getting records from database: %v\n", err)
		return nil
	}
	return records
}
