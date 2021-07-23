package boocat

import (
	"context"
	"errors"

	bcerrors "github.com/ivanmartinez/boocat/boocat/errors"
	"github.com/ivanmartinez/boocat/log"
)

// database is the database interface
type database interface {
	AddRecord(ctx context.Context, formatName string, record map[string]string) (string, error)
	UpdateRecord(ctx context.Context, formatName string, record map[string]string) error
	GetAllRecords(ctx context.Context, formatName string) ([]map[string]string, error)
	GetRecord(ctx context.Context, formatName, id string) (map[string]string, error)
	SearchRecord(ctx context.Context, formatName, value string) ([]map[string]string, error)
	ReferenceValidator(formatName string) Validate
}

type Boocat struct {
	formats map[string]Format
	db      database
}

func (bc *Boocat) SetDatabase(db database) *Boocat {
	bc.db = db
	return bc
}

func (bc *Boocat) SetFormat(name string, format Format) *Boocat {
	if bc.formats == nil {
		bc.formats = make(map[string]Format)
	}
	bc.formats[name] = format
	return bc
}

func (bc *Boocat) Formats() map[string]Format {
	return bc.formats
}

// TODO Don't log errors, return them

// GetRecord returns a record of a format (author, book...)
func (bc *Boocat) GetRecord(ctx context.Context, formatName string, id string) (map[string]string, error) {
	if bc.db == nil {
		log.Error.Println("database not set")
		return nil, bcerrors.InternalServerError{}
	}
	record, err := bc.db.GetRecord(ctx, formatName, id)
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
func (bc *Boocat) ListRecords(ctx context.Context, formatName string) ([]map[string]string, error) {
	if bc.db == nil {
		log.Error.Println("database not set")
		return nil, bcerrors.InternalServerError{}
	}
	records, err := bc.db.GetAllRecords(ctx, formatName)
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

// SearchRecords returns a slice of the records of a format (authors, books...) whose search fields contain the value
func (bc *Boocat) SearchRecords(ctx context.Context, formatName string, search string) ([]map[string]string, error) {
	if bc.db == nil {
		log.Error.Println("database not set")
		return nil, bcerrors.InternalServerError{}
	}
	records, err := bc.db.SearchRecord(ctx, formatName, search)
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
func (bc *Boocat) AddRecord(ctx context.Context, formatName string, record map[string]string) (string, error) {
	if bc.db == nil {
		log.Error.Println("database not set")
		return "", bcerrors.InternalServerError{}
	}
	format := bc.formats[formatName] // TODO: Check that the format exists
	failed := format.Validate(ctx, record)
	if len(failed) > 0 {
		return "", bcerrors.ValidationFailedError{Failed: failed}
	}
	id, err := bc.db.AddRecord(ctx, format.Name, record)
	switch {
	case errors.Is(err, bcerrors.ErrFormatNotFound):
		return "", bcerrors.ErrFormatNotFound
	case errors.Is(err, bcerrors.ErrRecordHasID):
		return "", bcerrors.ErrRecordHasID
	case err != nil:
		log.Error.Printf("adding record to database: %v\n", err)
		return "", bcerrors.InternalServerError{}
	}
	return id, nil
}

// UpdateRecord updates a record of a format (author, book...)
func (bc *Boocat) UpdateRecord(ctx context.Context, formatName string, record map[string]string) error {
	if bc.db == nil {
		log.Error.Println("database not set")
		return bcerrors.InternalServerError{}
	}
	format := bc.formats[formatName] // TODO: Check that the format exists
	failed := format.Validate(ctx, record)
	if len(failed) > 0 {
		return bcerrors.ValidationFailedError{Failed: failed}
	}
	err := bc.db.UpdateRecord(ctx, format.Name, record)
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
func (bc *Boocat) fillFromDatabase(ctx context.Context, record map[string]string, format Format) map[string]string {
	if bc.db == nil {
		log.Error.Println("database not set")
	}
	if format.IncompleteRecord(record) {
		if dbRecord, err := bc.db.GetRecord(ctx, format.Name, record["id"]); err == nil {
			record = format.Merge(record, dbRecord)
		} else {
			log.Error.Printf("getting record from database: %v\n", err)
		}
	}
	return record
}
