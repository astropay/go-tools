package datasource

import (
	"time"
)

// DBConfig holds the configuration required to initialize a data source
type DBConfig struct {
	DBName             string
	Driver             string
	Address            string
	Username           string
	Password           string
	MaxOpenConnections int
	MaxIdleConnections int
	ConnectionLifetime time.Duration
}
