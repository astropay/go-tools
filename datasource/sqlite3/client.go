package sqlite3

import (
	"github.com/jmoiron/sqlx"
	// import driver
	_ "github.com/mattn/go-sqlite3"
)

// Init connects to the database server
func Init(dbName string) (db *sqlx.DB, err error) {
	db, err = sqlx.Open("sqlite3", dbName)
	return
}

func RandFuncName() string {
	return "random()"
}
