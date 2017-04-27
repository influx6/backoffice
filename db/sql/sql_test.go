package sql_test

import (
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/influx6/backoffice/db"
	"github.com/influx6/backoffice/models/user"
	"github.com/influx6/faux/sink"
	"github.com/influx6/faux/sink/sinks"
	"github.com/influx6/faux/tests"
	"github.com/jmoiron/sqlx"
)

// contains different environment flags for use to setting up
// a db connection.
var (
	mydb djDB
	log  = sink.New(sinks.Stdout{})

	DBPortEnv     = "MYSQL_PORT"
	DBIPEnv       = "MYSQL_IP"
	DBUserEnv     = "MYSQL_USER"
	DBDatabaseEnv = "MYSQL_DATABASE"
	DBUserPassEnv = "MYSQL_PASSWORD"
)

type djDB struct{}

// New returns a new instance of a sqlx.DB connected to the db with the provided
// credentials pulled from the host environment.
func (djDB) New() (*sqlx.DB, error) {
	user := strings.TrimSpace(os.Getenv(DBUserEnv))
	userPass := strings.TrimSpace(os.Getenv(DBUserPassEnv))
	port := strings.TrimSpace(os.Getenv(DBPortEnv))
	ip := strings.TrimSpace(os.Getenv(DBIPEnv))
	dbName := strings.TrimSpace(os.Getenv(DBDatabaseEnv))

	if ip == "" {
		ip = "0.0.0.0"
	}

	addr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, userPass, ip, port, dbName)
	db, err := sqlx.Connect("mysql", addr)
	if err != nil {
		log.Emit(sinks.Error("Failed to connect to SQLServer: %+q", err).WithFields(sink.Fields{
			"addr":       addr,
			"mysql_ip":   ip,
			"mysql_port": port,
			"dbName":     dbName,
			"user":       user,
			"password":   userPass,
		}))

		return nil, err
	}

	return db, nil
}

func TestDBSave(t *testing.T) {
	nw, err := user.New(user.NewUser{
		Email:    "bob@guma.com",
		Password: "glow",
	})
	if err != nil {
		tests.Failed("Should have successfully created new user: %+q.", err)
	}
	tests.Passed("Should have successfully created new user.")

	dbInstance, err := mydb.New()
	if err != nil {
		tests.Failed("Should have successfully connected to mysql db: %+q.", err)
	}
	tests.Passed("Should have successfully conencted to mysql db.")

	if err := db.Save(log, dbInstance, nw); err != nil {
		tests.Failed("Should have successfully saved record to db table %q: %+q.", nw.Table(), err)
	}
	tests.Passed("Should have successfully saved record to db table %q: %+q.", nw.Table(), err)

	if err := db.Update(log, dbInstance, nw, "publid_id"); err != nil {
		tests.Failed("Should have successfully updated record to db table %q: %+q.", nw.Table(), err)
	}
	tests.Passed("Should have successfully updated record to db table %q: %+q.", nw.Table(), err)

	if err := db.Delete(log, dbInstance, nw, "public_id", nw.PublicID); err != nil {
		tests.Failed("Should have successfully deleted record to db table %q: %+q.", nw.Table(), err)
	}
	tests.Passed("Should have successfully deleted record to db table %q: %+q.", nw.Table(), err)
}

func getMainIP() string {
	udp, err := net.DialTimeout("udp", "8.8.8.8:80", 1*time.Millisecond)
	if err != nil {
		return "0.0.0.0"
	}

	defer udp.Close()

	localAddr := udp.LocalAddr().String()
	ip, _, _ := net.SplitHostPort(localAddr)

	return ip
}
