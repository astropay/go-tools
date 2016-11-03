package database

import "errors"

// DBType is used to internally represent generica database types
type DBType string

// Database types
const (
	DbTypeVarchar DBType = "VARCHAR"
	DbTypeNumeric DBType = "NUMERIC"
	DbTypeDate    DBType = "DATE"
	DbTypeBool    DBType = "BOOLEAN"
	// others...
)

// Database util errors
var (
	ErrInvalidFieldList = errors.New("Invalid field list")
)
