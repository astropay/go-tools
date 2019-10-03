package mysql

import (
	"fmt"

	// import MySQL driver
	"github.com/jmoiron/sqlx"
	_ "github.com/newrelic/go-agent/_integrations/nrmysql"
)

// Init connects to the database server
func Init(address, username, password, dbName string) (db *sqlx.DB, err error) {
	db, err = sqlx.Open("nrmysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=True", username, password, address, dbName))
	return
}

func RandFuncName() string {
	return "rand()"
}
