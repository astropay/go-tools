package mysql

import (
	"fmt"

	// import MySQL driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Init connects to the database server
func Init(address, username, password, dbName string) (db *sqlx.DB, err error) {
	db, err = sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=True", username, password, address, dbName))
	return
}

func RandFuncName() string {
	return "rand()"
}
