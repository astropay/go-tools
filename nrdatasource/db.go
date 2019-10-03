package nrdatasource

import (
	"errors"

	"github.com/astropay/go-tools/nrdatasource/mysql"
	"github.com/jmoiron/sqlx"
)

// Global errors
var (
	ErrNoDatabase         = errors.New("no database found with the specified name")
	ErrDriverNotSupported = errors.New("database driver not supported")
)

// DBAccess is the common interface for data access definitions
type DBAccess interface {
	New(config DBConfig) (*sqlx.DB, error)
	Get() (*sqlx.DB, error)
	Close() error
	CanLock() bool
	RandomFuncName() string
}

// Generic is the generic data access implementation for `DBAccess` interface.
// Drivers currently supported:
// - mysql
//
type Generic struct {
	db           *sqlx.DB
	canLock      bool
	randFuncName string
}

// New configures the datasources
func (g *Generic) New(config DBConfig) (db *sqlx.DB, err error) {
	switch config.Driver {

	case "mysql":
		db, err = mysql.Init(
			config.Address,
			config.Username,
			config.Password,
			config.DBName)
		if err == nil {
			db.SetMaxOpenConns(config.MaxOpenConnections)
			db.SetMaxIdleConns(config.MaxIdleConnections)
			db.SetConnMaxLifetime(config.ConnectionLifetime)
		} else {
			return nil, err
		}

		// ping DB to check if it's OK
		if err = db.Ping(); err != nil {
			return
		}

		// if exists a DB, close it
		if g.db != nil {
			g.db.Close()
		}

		g.db = db
		g.canLock = true
		g.randFuncName = mysql.RandFuncName()

	default:
		return nil, ErrDriverNotSupported
	}

	return
}

// Get returns the DB instance
func (g *Generic) Get() (db *sqlx.DB, err error) {
	if g.db != nil {
		db = g.db
	} else {
		err = ErrNoDatabase
	}
	return
}

// Close should be called when the server ends the execution,
// so connection are gracefully released
func (g *Generic) Close() (err error) {
	if g.db != nil {
		err = g.db.Close()
	}
	return
}

// CanLock returns true if the current driver supportes locking
func (g *Generic) CanLock() (lock bool) {
	if g != nil {
		lock = g.canLock
	}
	return
}

// RandomFuncName returns the driver's RANDOM name
func (g *Generic) RandomFuncName() (fName string) {
	if g != nil {
		fName = g.randFuncName
	}
	return
}
