package server

import (
	"context"
	"errors"

	"github.com/ivanmartinez/boocat/log"
	"github.com/ivanmartinez/boocat/server/database"
	bcerrors "github.com/ivanmartinez/boocat/server/errors"
	"github.com/ivanmartinez/boocat/server/formats"
)

var db database.DB

// Initialize initializes the server configuration without starting it. This is used in testing.
func Initialize(database database.DB) {
	db = database
}

// GetRecord returns a record of a format (author, book...)
func GetRecord(ctx context.Context, formatName string, id string) (map[string]string, error) {
	record, err := db.GetRecord(ctx, formatName, id)
	switch {
	case err == nil:
		return record, nil
	case errors.Is(err, bcerrors.ErrFormatNotFound):
		return nil, bcerrors.ErrFormatNotFound
	case errors.Is(err, bcerrors.ErrRecordNotFound):
		return nil, bcerrors.ErrRecordNotFound
	default:
		log.Error.Printf("getting record from database: %v\n", err)
		return nil, bcerrors.InternalServerError{}
	}
}

// ListRecords returns a slice of all records of a format (authors, books...)
func ListRecords(ctx context.Context, formatName string) ([]map[string]string, error) {
	records, err := db.GetAllRecords(ctx, formatName)
	switch {
	case err == nil:
		return records, nil
	case errors.Is(err, bcerrors.ErrFormatNotFound):
		return nil, bcerrors.ErrFormatNotFound
	default:
		log.Error.Printf("getting records from database: %v\n", err)
		return nil, bcerrors.InternalServerError{}
	}
}

// SearchRecords returns a slice of the records of a format (authors, books...) whose search fields contains the value
func SearchRecords(ctx context.Context, formatName string, search string) ([]map[string]string, error) {
	records, err := db.SearchRecord(ctx, formatName, search)
	switch {
	case err == nil:
		return records, nil
	case errors.Is(err, bcerrors.ErrFormatNotFound):
		return nil, bcerrors.ErrFormatNotFound
	default:
		log.Error.Printf("searching records from database: %v\n", err)
		return nil, bcerrors.InternalServerError{}
	}
}

// AddRecord adds a record of a format (author, book...)
func AddRecord(ctx context.Context, formatName string, record map[string]string) (string, error) {
	format := formats.Formats[formatName] // TODO: Check that the format exists
	failed := format.Validate(ctx, record)
	if len(failed) > 0 {
		return "", bcerrors.ValidationFailedError{Failed: failed}
	}
	id, err := db.AddRecord(ctx, format.Name, record)
	switch {
	case err == nil:
		return id, nil
	case errors.Is(err, bcerrors.ErrFormatNotFound):
		return "", bcerrors.ErrFormatNotFound
	case errors.Is(err, bcerrors.ErrRecordHasID):
		return "", bcerrors.ErrRecordHasID
	default:
		log.Error.Printf("adding record to database: %v\n", err)
		return "", bcerrors.InternalServerError{}
	}
}

// UpdateRecord updates a record of a format (author, book...)
func UpdateRecord(ctx context.Context, formatName string, record map[string]string) error {
	format := formats.Formats[formatName] // TODO: Check that the format exists
	failed := format.Validate(ctx, record)
	if len(failed) > 0 {
		return bcerrors.ValidationFailedError{Failed: failed}
	}
	err := db.UpdateRecord(ctx, format.Name, record)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, bcerrors.ErrFormatNotFound):
		return bcerrors.ErrFormatNotFound
	case errors.Is(err, bcerrors.ErrRecordDoesntHaveID):
		return bcerrors.ErrRecordDoesntHaveID
	default:
		log.Error.Printf("adding record to database: %v\n", err)
		return bcerrors.InternalServerError{}
	}
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
