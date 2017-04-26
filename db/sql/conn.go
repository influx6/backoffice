package sql

import (
	"fmt"

	"github.com/influx6/faux/sink"
	"github.com/influx6/faux/sink/sinks"
	"github.com/jmoiron/sqlx"
)

// Conn defines a struct which will generate a new sqlx connection for making
// sql queries.
type Conn struct {
	Port     int
	Addr     string
	User     string
	Driver   string
	Password string
	Database string
	Log      sink.Sink
}

// New returns a new instance of a sqlx.DB connected to the db with the provided
// credentials pulled from the host environment.
func (s Conn) New() (*sqlx.DB, error) {
	addr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", s.User, s.Password, s.Addr, s.Port, s.Addr)

	db, err := sqlx.Connect(s.Driver, addr)
	if err != nil {
		s.Log.Emit(sinks.Error("Failed to connect to sql server: %+q", err).WithFields(sink.Fields{
			"ip":     s.Addr,
			"port":   s.Port,
			"db":     s.Database,
			"user":   s.User,
			"driver": s.Driver,
		}))

		return nil, err
	}

	return db, nil
}
