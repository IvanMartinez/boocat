package sqlite

// SQLite implementation of the database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ivanmartinez/boocat/boocat"
	bcerrors "github.com/ivanmartinez/boocat/boocat/errors"
)

const (
	recordIDSize           = 20
	newRecordIDMaxAttempts = 5
	selectRecordIDSQL      = "SELECT record_id FROM field_values WHERE record_id=? LIMIT 1"
	insertFieldValuesSQL   = "INSERT INTO field_values (format, record_id, field, value) VALUES"
	updateFieldValueSQL    = "UPDATE field_values SET value=? WHERE format=? AND record_id=? AND field=?"
	selectRecordSQL        = "SELECT field, value FROM field_values WHERE format=? AND record_id=?"
	selectFormatRecordsSQL = "SELECT record_id, field, value FROM field_values WHERE format=?"
	driver                 = "sqlite3"
)

// sqliteDB stores data for this database interface
type sqliteDB struct {
	db *sql.DB
	selectRecordIDStmt,
	updateFieldValueStmt,
	selectRecordStmt,
	selectFormatRecordsStmt *sql.Stmt
}

// Ensure that sqliteDB implements boocat.Database
var _ boocat.Database = &sqliteDB{}

// Open opens the database connection and prepares the statements
func Open(ctx context.Context, dbDataSource *string) (*sqliteDB, error) {
	db, err := sql.Open(driver, *dbDataSource)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	// Enable foreign keys
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	sqliteDB := sqliteDB{db: db}
	sqliteDB.prepareStatements(ctx)
	return &sqliteDB, nil
}

// Close closes the statements and database connection
func (sdb *sqliteDB) Close() error {
	if err := sdb.selectRecordIDStmt.Close(); err != nil {
		return fmt.Errorf("closing selectRecordIDStmt: %v", err)
	}
	if err := sdb.updateFieldValueStmt.Close(); err != nil {
		return fmt.Errorf("closing updateFieldValueStmt: %v", err)
	}
	if err := sdb.selectRecordStmt.Close(); err != nil {
		return fmt.Errorf("closing selectRecordStmt: %v", err)
	}
	if err := sdb.selectFormatRecordsStmt.Close(); err != nil {
		return fmt.Errorf("closing selectFormatRecordsStmt: %v", err)
	}
	if err := sdb.db.Close(); err != nil {
		return fmt.Errorf("Closing DB: %v", err)
	}

	return nil
}

// AddRecord adds a new record of the format
func (sdb *sqliteDB) AddRecord(ctx context.Context, formatName string, record map[string]string) (newID string, err error) {
	if _, found := record["id"]; found {
		return "", bcerrors.ErrRecordHasID
	}
	err = sdb.inTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// Get a new record ID
		newRecordID, err := sdb.newRecordID(ctx, tx)
		if err != nil {
			return fmt.Errorf("getting new record ID: %w", err)
		}
		// Build the SQL INSERT statement and list of values
		const valueRow = "(?,?,?,?)"
		var valueRows []string
		vals := []interface{}{}
		for field, value := range record {
			valueRows = append(valueRows, valueRow)
			vals = append(vals, formatName, newRecordID, field, value)
		}
		insertSQL := insertFieldValuesSQL + strings.Join(valueRows, ",")
		// Execute the SQL statement
		res, err := tx.ExecContext(ctx, insertSQL, vals...)
		if err != nil {
			return fmt.Errorf("executing statement: %w", err)
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("getting affected rows after inserting: %w", err)
		}
		if rows != int64(len(record)) {
			return fmt.Errorf("inserted %d rows, should have been %d", rows, len(record))
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commiting transaction: %w", err)
		}
		return nil
	})
	return newID, err
}

// UpdateRecord updates a record of the format
func (sdb *sqliteDB) UpdateRecord(ctx context.Context, formatName string, record map[string]string) (err error) {
	id, found := record["id"]
	if !found {
		return bcerrors.ErrRecordDoesntHaveID
	}
	delete(record, "id")
	err = sdb.inTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		txUpdateBalanceStmt := tx.StmtContext(ctx, sdb.updateFieldValueStmt)
		for field, value := range record {
			// Execute the SQL statement
			res, err := txUpdateBalanceStmt.ExecContext(ctx, value, formatName, id, field)
			if err != nil {
				return fmt.Errorf("executing statement: %w", err)
			}
			rows, err := res.RowsAffected()
			if err != nil {
				return fmt.Errorf("getting affected rows after updating: %w", err)
			}
			if rows != 1 {
				return fmt.Errorf("updated %d rows, should have been 1", rows)
			}
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commiting transaction: %w", err)
		}
		return nil
	})
	return err
}

// GetRecord returns the record of the format with the id
func (sdb *sqliteDB) GetRecord(ctx context.Context, formatName, id string) (map[string]string, error) {
	rows, err := sdb.selectRecordStmt.QueryContext(ctx, formatName, id)
	if err != nil {
		return nil, fmt.Errorf("executing statement: %w", err)
	}
	record := make(map[string]string)
	for rows.Next() {
		var field, value string
		if err := rows.Scan(&field, &value); err != nil {
			return nil, fmt.Errorf("reading from rows: %w", err)
		}
		record[field] = value
	}
	record["id"] = id
	return record, nil
}

// GetAllRecords returns all records of a specific format from the database
func (sdb *sqliteDB) GetAllRecords(ctx context.Context, formatName string) ([]map[string]string, error) {
	rows, err := sdb.selectFormatRecordsStmt.QueryContext(ctx, formatName)
	if err != nil {
		return nil, fmt.Errorf("executing statement: %w", err)
	}
	recordFieldValue := make(map[string]map[string]string)
	for rows.Next() {
		var record_id, field, value string
		if err := rows.Scan(&record_id, &field, &value); err != nil {
			return nil, fmt.Errorf("reading from rows: %w", err)
		}
		if recordFieldValue[record_id] == nil {
			recordFieldValue[record_id] = make(map[string]string)
		}
		recordFieldValue[record_id][field] = value
	}
	records := make([]map[string]string, 0, len(recordFieldValue))
	for id, record := range recordFieldValue {
		record["id"] = id
		records = append(records, record)
	}
	return records, nil
}

// SearchRecord returns all records of the format that have the value in their searchable fields
func (sdb *sqliteDB) SearchRecord(ctx context.Context, formatName, value string) ([]map[string]string, error) {
	var documents []map[string]string
	return documentsToRecords(documents), nil
}

// ReferenceValidator returns a validator of references to records of the format
func (sdb *sqliteDB) ReferenceValidator(formatName string) boocat.Validate {
	return func(ctx context.Context, value interface{}) string {
		stringValue := fmt.Sprintf("%v", value)
		if _, err := sdb.GetRecord(ctx, formatName, stringValue); err != nil {
			return fmt.Sprintf("record of format '%s' and ID '%s' not found", formatName, stringValue)
		}
		return ""
	}
}

func (sdb *sqliteDB) newRecordID(ctx context.Context, tx *sql.Tx) (string, error) {
	for i := 0; i < newRecordIDMaxAttempts; i++ {
		newID := randomString(recordIDSize)
		txSelectRecordIDSQL := tx.StmtContext(ctx, sdb.selectRecordIDStmt)
		row := txSelectRecordIDSQL.QueryRowContext(ctx, newID)
		var existingID string
		err := row.Scan(&existingID)
		switch err {
		case sql.ErrNoRows:
			return newID, nil
		case nil:
			continue
		default:
			return "", fmt.Errorf("scanning row: %w", err)
		}
	}
	return "", fmt.Errorf("failed after %d attempts", newRecordIDMaxAttempts)
}

func (sdb *sqliteDB) inTransaction(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) error) (ferr error) {
	// Begin DB transaction
	tx, err := sdb.db.BeginTx(ctx, nil)
	// Rollback if not committed
	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			tx.Rollback()
			panic(p)
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ferr = fmt.Errorf("rolling transaction back: %w", err)
		}
	}()
	// Do fn
	return fn(ctx, tx)
}

func randomString(n int) string {
	var runes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}

// Return a slice of records from a slice of MongoDB documents. It just renames "_id" keys to "id".
func documentsToRecords(docs []map[string]string) []map[string]string {
	records := make([]map[string]string, 0, len(docs))
	for _, d := range docs {
		records = append(records, documentToRecord(d))
	}
	return records
}

// Returns a record from a MongoDB document. It just renames "_id" key to "id".
func documentToRecord(doc map[string]string) map[string]string {
	if id, found := doc["_id"]; found {
		delete(doc, "_id")
		doc["id"] = id
		return doc
	}
	return doc
}

// prepareStatements prepares the database SQL statements
func (sdb *sqliteDB) prepareStatements(ctx context.Context) {
	selectRecordIDStmt, err := sdb.db.PrepareContext(ctx, selectRecordIDSQL)
	if err != nil {
		log.Fatalf("couldn't prepare selectRecordIDStmt: %v", err)
	}
	sdb.selectRecordIDStmt = selectRecordIDStmt
	updateFieldValueStmt, err := sdb.db.PrepareContext(ctx, updateFieldValueSQL)
	if err != nil {
		log.Fatalf("couldn't prepare updateFieldValueStmt: %v", err)
	}
	sdb.updateFieldValueStmt = updateFieldValueStmt
	selectRecordStmt, err := sdb.db.PrepareContext(ctx, selectRecordSQL)
	if err != nil {
		log.Fatalf("couldn't prepare selectRecordStmt: %v", err)
	}
	sdb.selectRecordStmt = selectRecordStmt
	selectFormatRecordsStmt, err := sdb.db.PrepareContext(ctx, selectFormatRecordsSQL)
	if err != nil {
		log.Fatalf("couldn't prepare selectFormatRecordsStmt: %v", err)
	}
	sdb.selectFormatRecordsStmt = selectFormatRecordsStmt
}
