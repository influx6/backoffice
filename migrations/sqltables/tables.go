package sqltables

import (
	"github.com/influx6/backoffice/db/sql/tables"
	"github.com/influx6/faux/naming"
)

// BasicTables defines the migration table for creating the profiles's table.
func BasicTables(names naming.FeedNamer) []tables.TableMigration {
	var ts []tables.TableMigration

	ts = append(ts, tables.TableMigration{
		TableName:   names.New("profiles"),
		Timestamped: true,
		Indexes: []tables.IndexMigration{
			{
				IndexName: "user_id",
				Field:     "user_id",
			},
		},
		Fields: []tables.FieldMigration{
			{
				FieldName: "user_id",
				FieldType: "VARCHAR(255)",
				NotNull:   true,
			},
			{
				FieldName: "address",
				FieldType: "text",
				NotNull:   true,
			},
			{
				FieldName:  "public_id",
				FieldType:  "VARCHAR(255)",
				PrimaryKey: true,
				NotNull:    true,
			},
			{
				FieldName: "first_name",
				FieldType: "VARCHAR(255)",
				NotNull:   true,
			},
			{
				FieldName: "last_name",
				FieldType: "VARCHAR(255)",
				NotNull:   true,
			},
		},
	})

	ts = append(ts, tables.TableMigration{
		TableName:   names.New("sessions"),
		Timestamped: true,
		Indexes: []tables.IndexMigration{
			{
				IndexName: "user_id",
				Field:     "user_id",
			},
		},
		Fields: []tables.FieldMigration{
			{
				FieldName: "user_id",
				FieldType: "VARCHAR(255)",
				NotNull:   true,
			},
			{
				FieldName: "token",
				FieldType: "VARCHAR(255)",
				NotNull:   true,
			},
			{
				FieldName:  "public_id",
				FieldType:  "VARCHAR(255)",
				PrimaryKey: true,
				NotNull:    true,
			},
			{
				FieldName: "expires",
				FieldType: "timestamp",
				NotNull:   true,
			},
		},
	})

	ts = append(ts, tables.TableMigration{
		TableName:   names.New("users"),
		Timestamped: true,
		Indexes:     []tables.IndexMigration{},
		Fields: []tables.FieldMigration{
			{
				FieldName: "email",
				FieldType: "VARCHAR(255)",
				NotNull:   true,
			},
			{
				FieldName:  "public_id",
				FieldType:  "VARCHAR(255)",
				PrimaryKey: true,
				NotNull:    true,
			},
			{
				FieldName: "private_id",
				FieldType: "VARCHAR(255)",
				NotNull:   true,
			},
			{
				FieldName: "hash",
				FieldType: "VARCHAR(255)",
				NotNull:   true,
			},
		},
	})

	return ts
}
