package sql

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/influx6/backoffice/db"
	"github.com/influx6/backoffice/db/sql/tables"
	"github.com/influx6/faux/sink"
	"github.com/influx6/faux/sink/sinks"
	"github.com/jmoiron/sqlx"
)

// contains templates of sql statement for use in operations.
const (
	countTemplate         = "SELECT %s FROM %s"
	selectAllTemplate     = "SELECT * FROM %s ORDER BY %s %s"
	selectLimitedTemplate = "SELECT * FROM %s ORDER BY %s %s LIMIT %d %d"
	selectItemTemplate    = "SELECT * FROM %s WHERE %s=?"
	insertTemplate        = "INSERT INTO %s %s VALUES %s"
	updateTemplate        = "UPDATE %s SET %s WHERE %s=%s"
	deleteTemplate        = "DELETE FROM %s WHERE %s=?"
)

// DB defines an interface which exposes a method to return a new
// underline sql.Database.
type DB interface {
	New() (*sqlx.DB, error)
}

// SQL defines an struct which implements the db.Provider which allows us
// execute CRUD ops.
type SQL struct {
	d      DB
	l      sink.Sink
	inited bool
	tables []tables.TableMigration
}

// New returns a new instance of SQL.
func New(s sink.Sink, d DB, ts ...tables.TableMigration) *SQL {
	return &SQL{
		d:      d,
		l:      s,
		tables: ts,
	}
}

// migrate takes the individual query supplied and attempts to
// execute them returning any error found.
func (sq *SQL) migrate() error {
	if sq.d == nil {
		return nil
	}

	if sq.inited {
		return nil
	}

	dbi, err := sq.d.New()
	if err != nil {
		return err
	}

	defer dbi.Close()

	for _, table := range sq.tables {
		sq.l.Emit(sinks.Info("Executing Migration").WithFields(sink.Fields{
			"query": table.String(),
			"table": table.TableName,
		}))

		if _, err := dbi.Exec(table.String()); err != nil {
			sq.l.Emit(sinks.Error(err).WithFields(sink.Fields{"query": table.String(), "table": table.TableName}))
			return err
		}
	}

	sq.inited = true

	return nil
}

// Save takes the giving table name with the giving fields and attempts to save this giving
// data appropriately into the giving db.
func (sq *SQL) Save(log sink.Sink, identity db.TableIdentity, table db.TableFields) error {
	defer log.Emit(sinks.Info("Save to DB").With("table", identity.Table()).Trace("db.Save").End())

	if err := sq.migrate(); err != nil {
		return err
	}

	db, err := sq.d.New()
	if err != nil {
		return err
	}

	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	fields := table.Fields()
	fieldNames := FieldNames(fields)
	values := FieldValues(fieldNames, fields)

	fieldNames = append(fieldNames, "created_at")
	fieldNames = append(fieldNames, "updated_at")

	values = append(values, time.Now())
	values = append(values, time.Now())

	query := fmt.Sprintf(insertTemplate, identity.Table(), FieldNameMarkers(fieldNames), FieldMarkers(len(fieldNames)))
	log.Emit(sinks.Info("DB:Query").With("query", query))

	if _, err := db.Exec(query, values...); err != nil {
		log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"err":   err,
			"query": query,
			"table": identity.Table(),
		}))
		return err
	}

	return tx.Commit()
}

// Update takes the giving table name with the giving fields and attempts to update this giving
// data appropriately into the giving db.
// index - defines the string which should identify the key to be retrieved from the fields to target the
// data to be updated in the db.
func (sq *SQL) Update(log sink.Sink, identity db.TableIdentity, table db.TableFields, index string) error {
	defer log.Emit(sinks.Info("Update to DB").With("table", identity.Table()).Trace("db.Update").End())

	if err := sq.migrate(); err != nil {
		return err
	}

	db, err := sq.d.New()
	if err != nil {
		return err
	}

	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	tableFields := table.Fields()
	tableFields["updated_at"] = time.Now()

	// Given index was not found, return error.
	indexValue, ok := tableFields[index]
	if !ok {
		return errors.New("Index key not found in fields")
	}

	// Delete given index from fieldNames
	delete(tableFields, index)

	fieldNames := FieldNamesFromMap(tableFields)

	query := fmt.Sprintf(updateTemplate, identity.Table(), FieldMarkers(len(fieldNames)), index, indexValue)
	log.Emit(sinks.Info("DB:Query").With("query", query))

	if _, err := db.Exec(query, FieldValues(fieldNames, tableFields)...); err != nil {
		log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"err":   err,
			"query": query,
			"table": identity.Table(),
		}))
		return err
	}

	return tx.Commit()
}

// GetAllPerPage retrieves the giving data from the specific db with the specific index and value.
func (sq *SQL) GetAllPerPage(log sink.Sink, table db.TableIdentity, order string, orderBy string, page int, responserPerPage int) ([]map[string]interface{}, int, error) {
	defer log.Emit(sinks.Info("Retrieve all records from DB").With("table", table.Table()).WithFields(sink.Fields{
		"order":            order,
		"page":             page,
		"responserPerPage": responserPerPage,
	}).Trace("db.GetAll").End())

	if err := sq.migrate(); err != nil {
		return nil, 0, err
	}

	db, err := sq.d.New()
	if err != nil {
		return nil, 0, err
	}

	defer db.Close()

	if page < 0 && responserPerPage < 0 {
		records, err := sq.GetAll(log, table, order, orderBy)
		return records, len(records), err
	}

	// Get total number of records.
	totalRecords, err := sq.Count(log, table, "public_id")
	if err != nil {
		return nil, 0, err
	}

	switch strings.ToLower(order) {
	case "asc":
		order = "ASC"
	case "dsc", "desc":
		order = "DESC"
	default:
		order = "ASC"
	}

	var fields []map[string]interface{}

	var totalWanted, indexToStart int

	if page < 0 && responserPerPage > 0 {
		totalWanted = responserPerPage
		indexToStart = 0
	} else {
		totalWanted = responserPerPage * page
		indexToStart = totalWanted / 2

		if page > 1 {
			indexToStart++
		}
	}

	// If we are passed the total, just return nil records and total without error.
	if indexToStart > totalRecords {
		return nil, totalRecords, nil
	}

	query := fmt.Sprintf(selectLimitedTemplate, table.Table(), orderBy, order, indexToStart, totalWanted)
	log.Emit(sinks.Info("DB:Query:GetAllPerPage").With("query", query))

	if err := db.Select(&fields, query); err != nil {
		log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"err":   err,
			"query": query,
			"table": table.Table(),
		}))

		return nil, totalRecords, err
	}

	return fields, totalRecords, nil
}

// GetAll retrieves the giving data from the specific db with the specific index and value.
func (sq *SQL) GetAll(log sink.Sink, table db.TableIdentity, order string, orderBy string) ([]map[string]interface{}, error) {
	defer log.Emit(sinks.Info("Retrieve all records from DB").With("table", table.Table()).Trace("db.GetAll").End())

	if err := sq.migrate(); err != nil {
		return nil, err
	}

	db, err := sq.d.New()
	if err != nil {
		return nil, err
	}

	defer db.Close()

	switch strings.ToLower(order) {
	case "asc":
		order = "ASC"
	case "dsc", "desc":
		order = "DESC"
	default:
		order = "ASC"
	}

	var fields []map[string]interface{}

	query := fmt.Sprintf(selectAllTemplate, table.Table(), orderBy, order)
	log.Emit(sinks.Info("DB:Query:GetAll").With("query", query))

	rows, err := db.Queryx(query)
	if err != nil {
		log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"err":   err,
			"query": query,
			"table": table.Table(),
		}))
		return nil, err
	}

	for rows.Next() {
		mo := make(map[string]interface{})
		if err := rows.MapScan(mo); err != nil {
			log.Emit(sinks.Error(err).WithFields(sink.Fields{
				"err":   err,
				"query": query,
				"table": table.Table(),
			}))
			return nil, err
		}

		fields = append(fields, mo)
	}

	return fields, nil
}

// Get retrieves the giving data from the specific db with the specific index and value.
func (sq *SQL) Get(log sink.Sink, table db.TableIdentity, consumer db.TableConsumer, index string, indexValue interface{}) error {
	defer log.Emit(sinks.Info("Get record from DB").WithFields(sink.Fields{
		"table":      table.Table(),
		"index":      index,
		"indexValue": indexValue,
	}).Trace("db.Get").End())

	if err := sq.migrate(); err != nil {
		return err
	}

	db, err := sq.d.New()
	if err != nil {
		return err
	}

	defer db.Close()

	query := fmt.Sprintf(selectItemTemplate, table.Table(), index)
	log.Emit(sinks.Info("DB:Query").With("query", query))

	row := db.QueryRowx(query, indexValue)
	if err := row.Err(); err != nil {
		log.Emit(sinks.Error("DB:Query: %+q", err).WithFields(sink.Fields{
			"err":   err,
			"query": query,
			"table": table.Table(),
		}))
		return err
	}

	mo := make(map[string]interface{})

	if err := row.MapScan(mo); err != nil {
		log.Emit(sinks.Error("DB:Query: %+q", err).WithFields(sink.Fields{
			"err":   err,
			"query": query,
			"table": table.Table(),
		}))

		return err
	}

	consumer.WithFields(mo)

	return nil
}

// Count retrieves the total number of records from the specific table from the db.
func (sq *SQL) Count(log sink.Sink, table db.TableIdentity, index string) (int, error) {
	defer log.Emit(sinks.Info("Count record from DB").WithFields(sink.Fields{
		"table": table.Table(),
	}).Trace("db.Get").End())

	if err := sq.migrate(); err != nil {
		return 0, err
	}

	db, err := sq.d.New()
	if err != nil {
		return 0, err
	}

	defer db.Close()

	var records []interface{}

	query := fmt.Sprintf(countTemplate, index, table.Table())
	log.Emit(sinks.Info("DB:Query").With("query", query))

	if err := db.Select(&records, query); err != nil {
		log.Emit(sinks.Error("DB:Query").WithFields(sink.Fields{
			"err":   err,
			"query": query,
			"table": table.Table(),
		}))
		return 0, err
	}

	return len(records), nil
}

// Delete removes the giving data from the specific db with the specific index and value.
func (sq *SQL) Delete(log sink.Sink, table db.TableIdentity, index string, indexValue interface{}) error {
	defer log.Emit(sinks.Info("Delete record from DB").WithFields(sink.Fields{
		"table":      table.Table(),
		"index":      index,
		"indexValue": indexValue,
	}).Trace("db.GetAll").End())

	if err := sq.migrate(); err != nil {
		return err
	}

	db, err := sq.d.New()
	if err != nil {
		return err
	}

	defer db.Close()

	var fields map[string]interface{}

	query := fmt.Sprintf(deleteTemplate, table.Table(), index)
	log.Emit(sinks.Info("DB:Query").With("query", query))

	if err := db.Get(&fields, query, indexValue); err != nil {
		log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"err":   err,
			"query": query,
			"table": table.Table(),
		}))
		return err
	}

	return nil
}

// FieldMarkers returns a (?,...,>) string which represents
// all filedNames extrated from the provided TableField.
func FieldMarkers(total int) string {
	var markers []string

	for i := 0; i < total; i++ {
		markers = append(markers, "?")
	}

	return "(" + strings.Join(markers, ",") + ")"
}

// FieldNameMarkers returns a (fieldName,...,fieldName) string which represents
// all filedNames extrated from the provided TableField.
func FieldNameMarkers(fields []string) string {
	return "(" + strings.Join(fields, ", ") + ")"
}

// FieldValues returns a (fieldName,...,fieldName) string which represents
// all filedNames extrated from the provided TableField.
func FieldValues(names []string, fields map[string]interface{}) []interface{} {
	var vals []interface{}

	for _, name := range names {
		vals = append(vals, fields[name])
	}

	return vals
}

// FieldNamesFromMap returns a (fieldName,...,fieldName) string which represents
// all filedNames extrated from the provided TableField.
func FieldNamesFromMap(fields map[string]interface{}) []string {
	var names []string

	for key := range fields {
		names = append(names, key)
	}

	return names
}

// FieldNames returns a (fieldName,...,fieldName) string which represents
// all filedNames extrated from the provided TableField.
func FieldNames(fields map[string]interface{}) []string {
	var names []string

	for key := range fields {
		names = append(names, key)
	}

	return names
}
