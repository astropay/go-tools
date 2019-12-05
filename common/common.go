/*
	@author Robert
*/

package common

import (
	"errors"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unsafe"
)

var (
	letterAndNumberRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	// email validation regular expression
	emailRegEx = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

	randomGenerator = rand.New(rand.NewSource(time.Now().Unix()))
)

// IsEmailAddress returns true if str seems to be an email address
func IsEmailAddress(str string) bool {
	return emailRegEx.MatchString(str)
}

// UpdateStructFields() errors
var (
	ErrNotStruct = errors.New("destination must by struct or a pointer to struct")
)

// UpdateStructFromMap can be used to update the fields of a structure by sending
// the new field values in a map, wherer the key is the field name as in the struct
// and the value is the new value that will be set.
//
// Only those values in the map will be evaluated and updated.
// Parameter destination must be a pointer, otherwise changes won't be reflected.
func UpdateStructFromMap(destination interface{}, source map[string]interface{}) (err error) {

	// to avoid panic, type of 'destination' must be a struct or a pointer
	if reflect.TypeOf(destination).Kind() == reflect.Struct || reflect.TypeOf(destination).Kind() == reflect.Ptr {
		if ps := reflect.ValueOf(destination); ps.IsValid() {
			s := ps.Elem()

			if s.Kind() == reflect.Struct {
				for k, v := range source {
					// retrieve field from struct
					if field := s.FieldByName(k); field.IsValid() {

						if field.CanSet() {

							// set field value based on the type
							// note: not all types are supported, so test and add as needed
							switch field.Kind() {

							case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
								field.SetInt(v.(int64))
							case reflect.String:
								field.SetString(v.(string))
							case reflect.Bool:
								field.SetBool(v.(bool))
							case reflect.Float32, reflect.Float64:
								field.SetFloat(v.(float64))
							case reflect.Ptr:
								field.SetPointer(v.(unsafe.Pointer))
							case reflect.Complex64, reflect.Complex128:
								field.SetComplex(v.(complex128))
							}
						}
					}
				}
			} else {
				err = ErrNotStruct
			}
		}
	} else {
		err = ErrNotStruct
	}

	return
}

// Random generates a random number between min and max.
func Random(min, max int) int {
	return randomGenerator.Intn(max-min) + min
}

// RandomString generates a random string of the specified length.
// Keep in mind that random seed must be initialized before. Example:
// 		rand.Seed(time.Now().Unix())
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterAndNumberRunes[randomGenerator.Intn(len(letterAndNumberRunes))]
	}
	return string(b)
}

// MaskString creates a mask with `maskChar` for the indicated string `s`.
// If noMaskLeft and noMaskRight equals -1, then all the string is masked.
func MaskString(s string, noMaskLeft, noMaskRight int, maskChar string) (masked string) {

	// return all masked string if applies
	if len(s) <= noMaskLeft+noMaskRight || (noMaskLeft == -1 && noMaskRight == -1) {
		masked = strings.Repeat(maskChar, len(s))
		return
	}

	if noMaskLeft == -1 {
		noMaskLeft = 0
	}

	if noMaskRight == -1 {
		noMaskRight = 0
	}

	sLen := len(s)

	leftStr := s[:noMaskLeft]
	rightStr := s[sLen-noMaskRight:]
	middle := strings.Repeat(maskChar, sLen-len(leftStr)-len(rightStr))

	masked = leftStr + middle + rightStr
	return
}

// FindInStringArray searches for the indicated element 'elem' in array 'a'
func FindInStringArray(elem string, a []string) (e string, found bool) {
	for i := range a {
		if a[i] == elem {
			return elem, true
		}
	}

	return "", false
}

// FindInIntgArray searches for the indicated element 'elem' in array 'a'
func FindInIntgArray(elem int64, a []int64) (e int64, found bool) {
	for i := range a {
		if a[i] == elem {
			return elem, true
		}
	}

	return 0, false
}

// GetNonNullFields returns an array with all the fields that
// aren't nil in the structure's instance
func GetNonNullFields(i interface{}, tagName string) (fields []string) {

	var e reflect.Value
	v := reflect.ValueOf(i)

	if v.Kind() == reflect.Struct {
		e = v
	} else if v.Kind() == reflect.Ptr {
		e = v.Elem()
	} else {
		// non applicable
		return
	}

	for f := 0; f < e.NumField(); f++ {
		process := false
		fieldInstance := e.Type().Field(f)
		fieldKind := e.Type().Kind()

		// skip structs and pointers to structs
		switch fieldKind {
		case reflect.Ptr:
			if e.Field(f).Elem().Kind() != reflect.Struct {
				process = true
			}
		default:
			process = true
		}

		if process {
			if !e.Field(f).IsNil() {
				if tagName != "" {
					if jsonTag := fieldInstance.Tag.Get(tagName); jsonTag != "" && jsonTag != "-" {
						fieldName := strings.Split(jsonTag, ",")[0]
						fields = append(fields, fieldName)
					}
				} else {
					fields = append(fields, fieldInstance.Name)
				}
			}
		}
	}

	return
}
