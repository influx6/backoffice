package sql_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/influx6/backoffice/db"
	"github.com/influx6/backoffice/db/sql"
	"github.com/influx6/backoffice/migrations/sqltables"
	"github.com/influx6/backoffice/models/user"
	"github.com/influx6/faux/naming"
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

func TestSQLAPI(t *testing.T) {
	basicNamer := naming.NewNamer("%s_%s", naming.PrefixNamer{Prefix: "test"})
	userTable := db.TableName{Name: basicNamer.New("users")}

	nw, err := user.New(user.NewUser{
		Email:    "bob@guma.com",
		Password: "glow",
	})
	if err != nil {
		tests.Failed("Should have successfully created new user: %+q.", err)
	}
	tests.Passed("Should have successfully created new user.")

	db := sql.New(log, mydb, sqltables.BasicTables(basicNamer)...)

	t.Logf("Given the need to validate sql api operations")
	{

		t.Log("\tWhen saving user record")
		{
			if err := db.Save(userTable, nw); err != nil {
				tests.Failed("Should have successfully saved record to db table %q: %+q.", userTable.Table(), err)
			}
			tests.Passed("Should have successfully saved record to db table %q.", userTable.Table())
		}

		t.Log("\tWhen counting user records")
		{
			total, err := db.Count(userTable)
			if err != nil {
				tests.Failed("Should have successfully saved record to db table %q: %+q.", userTable.Table(), err)
			}
			tests.Passed("Should have successfully saved record to db table %q.", userTable.Table())

			if total <= 0 {
				tests.Failed("Should have successfully recieved a count greater than 0.")

			}
			tests.Passed("Should have successfully recieved a count greater than 0.")
		}

		t.Log("\tWhen retrieving all user record")
		{
			records, err := db.GetAll(userTable, "asc", "public_id")
			if err != nil {
				tests.Failed("Should have successfully retrieved all records from db table %q: %+q.", userTable.Table(), err)
			}
			tests.Passed("Should have successfully retrieved all records from db table %q.", userTable.Table())

			if len(records) == 0 {
				tests.Failed("Should have successfully retrieved atleast one record from db table %q.", userTable.Table())
			}
			tests.Passed("Should have successfully retrieved atleast one record from db table %q.", userTable.Table())
		}

		t.Log("\tWhen retrieving all user record based on page")
		{
			_, total, err := db.GetAllPerPage(userTable, "asc", "public_id", 2, 2)
			if err != nil {
				tests.Failed("Should have successfully retrieved all records from db table %q: %+q.", userTable.Table(), err)
			}
			tests.Passed("Should have successfully retrieved all records from db table %q.", userTable.Table())

			if total == -1 {
				tests.Failed("Should have successfully retrieved records based on pages from db table %q.", userTable.Table())
			}
			tests.Passed("Should have successfully retrieved records based on pages from db table %q.", userTable.Table())
		}

		t.Log("\tWhen retrieving user record")
		{
			var nu user.User
			if err := db.Get(userTable, &nu, "public_id", nw.PublicID); err != nil {
				tests.Failed("Should have successfully retrieved record from db table %q: %+q.", userTable.Table(), err)
			}
			tests.Passed("Should have successfully retrieved record from db table %q.", userTable.Table())

			if nu.PublicID != nw.PublicID {
				tests.Info("Expected: %+q", nw.Fields())
				tests.Info("Recieved: %+q", nu.Fields())
				tests.Failed("Should have successfully matched original user with user retrieved from db.")
			}
			tests.Passed("Should have successfully matched original user with user retrieved from db.")
		}

		t.Log("\tWhen updating user record")
		{
			if err := db.Update(userTable, nw, "public_id"); err != nil {
				tests.Failed("Should have successfully updated record to db table %q: %+q.", userTable.Table(), err)
			}
			tests.Passed("Should have successfully updated record to db table %q.", userTable.Table())
		}

		t.Logf("\tWhen deleting user record")
		{
			if err := db.Delete(userTable, "public_id", nw.PublicID); err != nil {
				tests.Failed("Should have successfully deleted record to db table %q: %+q.", userTable.Table(), err)
			}
			tests.Passed("Should have successfully deleted record to db table %q.", userTable.Table())
		}
	}
}
