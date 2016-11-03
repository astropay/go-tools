package database

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

// BuildUpdateSetQuery returns a string that can be used to build a set query, with
// the values put as part of the string
func BuildUpdateSetQuery(obj interface{}, fields []string) (string, error) {

	checkType := reflect.TypeOf(obj)

	// obj must be struct or pointer to struct
	if checkType.Kind() != reflect.Ptr && checkType.Kind() != reflect.Struct {
		return "", fmt.Errorf("Invalid obj type '%s'", checkType.Kind().String())
	}

	var objVal reflect.Value
	var objType reflect.Type

	if checkType.Kind() == reflect.Ptr {
		objVal = reflect.ValueOf(obj).Elem()
		objType = checkType.Elem()
	} else {
		objVal = reflect.ValueOf(obj)
		objType = checkType
	}

	if fields != nil && len(fields) > 0 {

		// prepare sql
		buf := new(bytes.Buffer)
		buf.WriteString("SET ")

		for i := 0; i < len(fields); i++ {
			var (
				field  reflect.StructField
				exists bool
			)

			if field, exists = objType.FieldByName(fields[i]); exists {

				fieldInstance := objVal.FieldByName(fields[i])
				colName := resolveColumnName(field)
				colType := resolveColumnType(field)

				// get field value
				fieldKind := fieldInstance.Kind()
				var fieldValue interface{}

				if fieldKind == reflect.Ptr {
					fieldValue = fieldInstance.Elem().Interface()
				} else {
					fieldValue = fieldInstance.Interface()
				}

				// create field=value string
				buf.WriteString(fmt.Sprintf(buildFieldSet(colType), colName, fieldValue))
			} else {
				return "", fmt.Errorf("Invalid field '%s'", fields[i])
			}

			// add separator
			if exists && i < len(fields)-1 {
				buf.WriteString(",")
			}
		}

		return buf.String(), nil
	}

	return "", ErrInvalidFieldList
}

// BuildParametrizedUpdateSetQuery returns a string that can be used to build a set query,
// using parameters instead of values (parameter used is '?')
func BuildParametrizedUpdateSetQuery(obj interface{}, fields []string) (string, error) {

	checkType := reflect.TypeOf(obj)

	// obj must be struct or pointer to struct
	if checkType.Kind() != reflect.Ptr && checkType.Kind() != reflect.Struct {
		return "", fmt.Errorf("Invalid obj type '%s'", checkType.Kind().String())
	}

	var objType reflect.Type
	if checkType.Kind() == reflect.Ptr {
		objType = checkType.Elem()
	} else {
		objType = checkType
	}

	if fields != nil && len(fields) > 0 {

		// prepare sql
		buf := new(bytes.Buffer)
		buf.WriteString("SET ")

		for i := 0; i < len(fields); i++ {
			var (
				field  reflect.StructField
				exists bool
			)

			if field, exists = objType.FieldByName(fields[i]); exists {
				colName := resolveColumnName(field)
				buf.WriteString(fmt.Sprintf("%s=?", colName))
			} else {
				return "", fmt.Errorf("Invalid field '%s'", fields[i])
			}

			// add separator
			if exists && i < len(fields)-1 {
				buf.WriteString(",")
			}
		}

		return buf.String(), nil
	}

	return "", ErrInvalidFieldList
}

// GetParameterValues returns an array that may be used in a parametrized query
func GetParameterValues(obj interface{}, fields []string, args ...interface{}) ([]interface{}, error) {

	checkType := reflect.TypeOf(obj)

	// obj must be struct or pointer to struct
	if checkType.Kind() != reflect.Ptr && checkType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Invalid obj type '%s'", checkType.Kind().String())
	}

	var objVal reflect.Value
	var objType reflect.Type

	if checkType.Kind() == reflect.Ptr {
		objVal = reflect.ValueOf(obj).Elem()
		objType = checkType.Elem()
	} else {
		objVal = reflect.ValueOf(obj)
		objType = checkType
	}

	if fields != nil && len(fields) > 0 {

		// check if there are other arguments to include; those arguments will
		// be added in the end (good for a where condition)
		var otherFields int
		if args != nil {
			otherFields = len(args)
		}

		params := make([]interface{}, len(fields)+otherFields)

		for i := 0; i < len(fields); i++ {
			if _, exists := objType.FieldByName(fields[i]); exists {

				// get field value
				var fieldValue interface{}
				fieldInstance := objVal.FieldByName(fields[i])
				fieldKind := fieldInstance.Kind()

				if fieldKind == reflect.Ptr {
					fieldValue = fieldInstance.Elem().Interface()
				} else {
					fieldValue = fieldInstance.Interface()
				}

				// add field value to array
				params[i] = fieldValue
			} else {
				return nil, fmt.Errorf("Invalid field '%s'", fields[i])
			}
		}

		// other arguments?
		if otherFields > 0 {
			for i := len(fields); i < len(fields)+otherFields; i++ {
				idx := i - len(fields)
				params[i] = args[idx]
			}
		}

		return params, nil
	}

	return nil, ErrInvalidFieldList
}

func buildFieldSet(dbType DBType) string {
	// format strings used to build sentences
	quotedFormat := "`%s`='%v'"
	unquotedFormat := "`%s`=%v"

	switch dbType {
	case DbTypeVarchar, DbTypeDate:
		return quotedFormat
	case DbTypeBool, DbTypeNumeric:
		return unquotedFormat
	default:
		return unquotedFormat
	}
}

// resolves the column name associated to a struct's field;
// tag 'db' is used for compatibility with "github.com/jmoiron/sqlx"
func resolveColumnName(field reflect.StructField) (col string) {
	if dbColumn := field.Tag.Get("db"); dbColumn != "" {
		col = dbColumn
	} else {
		// TODO: it would be great to use camel case for field names and 'automagically'
		//       convert something like 'FieldName' to 'field_name'; but it's just an idea :-)
		col = strings.ToLower(field.Name)
	}

	return
}

// resolves the column data type associated to a field
func resolveColumnType(field reflect.StructField) (dbType DBType) {
	if fieldType := field.Tag.Get("db_type"); fieldType != "" {
		// db type was defined in the tag
		dbType = DBType(strings.ToUpper(fieldType))
	} else {
		// get type from struct definition
		dbType = mapKindToDBType(field.Type.Kind())
	}

	return
}

// maps the kind of the field to out internal type representation;
// it used only to wrap (or not) the field value between quotes ('value')
func mapKindToDBType(kind reflect.Kind) (dbType DBType) {
	switch kind {
	case reflect.String:
		dbType = DbTypeVarchar
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int, reflect.Float32, reflect.Float64:
		dbType = DbTypeNumeric
	case reflect.Bool:
		dbType = DbTypeBool
	default:
		dbType = DbTypeVarchar
	}

	return
}
