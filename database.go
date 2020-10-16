package main

import (
	"log"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

func init() {
	// this Pings the database trying to connect, panics on error
	// use sqlx.Open() for sql.Open() semantics
	var err error
	db, err = sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		sentry.CaptureException(err)
		log.Fatalln(err) // TODO: Report this error to sentry?!?
	}

	// exec the schema or fail; multi-statement Exec behavior varies between
	// database drivers;  pq will exec them all, sqlite3 won't, ymmv
	// db.MustExec(schema)
}
