package db

import (
	"github.com/influx6/faux/sink"
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

// TableIdentity defines an interface which exposes a method returning table name
// associated with the giving implementing structure.
type TableIdentity interface {
	Table() string
}

// TableFields defines an interface which exposes method to return a map of all data
// associated with the defined structure.
type TableFields interface {
	Fields() map[string]interface{}
}

// TableConsumer defines an interface which accepts a map of data which will be loaded
// into the giving implementing structure.
type TableConsumer interface {
	WithFields(map[string]interface{}) error
}

// Migration defines an interface which provides structures to setup a new db migration
// call.
type Migration interface {
	Migrate() error
}

// DB defines a type which allows CRUD operations provided by a underline
// db structure.
type DB interface {
	Save(s sink.Sink, t TableIdentity, f TableFields) error
	Count(s sink.Sink, t TableIdentity, index string) (int, error)
	Update(s sink.Sink, t TableIdentity, f TableFields, index string) error
	Delete(s sink.Sink, t TableIdentity, index string, value interface{}) error
	Get(s sink.Sink, t TableIdentity, c TableConsumer, index string, value interface{}) error
	GetAll(s sink.Sink, t TableIdentity, order string, orderBy string) ([]map[string]interface{}, error)
	GetAllPerPage(s sink.Sink, t TableIdentity, order string, orderBy string, page int, responserPage int) ([]map[string]interface{}, int, error)
}

//=============================================================================================================================================

// TableName defines a struct which returns a given table name associated with the table.
type TableName struct {
	Name string
}

// Table returns the giving name associated with the struct.
func (t TableName) Table() string {
	return t.Name
}

//=============================================================================================================================================

// Namer exposes a method which can be used to generate specific names based on a giving
// critieras.
type Namer interface {
	New(string) string
}

//=============================================================================================================================================

// TableNamer defines holds a underline naming mechanism to deliver new TableName instance.
type TableNamer struct {
	namer Namer
}

// NewTableNamer returns a new TableNamer instance.
func NewTableNamer(nm Namer) *TableNamer {
	return &TableNamer{
		namer: nm,
	}
}

// New returns a new TableName which is fed into the underline naming mechanism to
// generate a unique name for that table.
// eg namer = FeedNamer("sugo_company");  TableNamer(namer).New("users") => "sugo_company_users".
func (t *TableNamer) New(table string) TableName {
	return TableName{Name: t.namer.New(table)}
}

//=============================================================================================================================================
