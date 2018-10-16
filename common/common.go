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
	"unsafe"
)

var letterAndNumberRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

// IsEmailAddress returns true if str seems to be an email address
func IsEmailAddress(str string) bool {
	reg := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return reg.MatchString(str)
}

// UpdateStructFields() errors
var (
	ErrNotStruct = errors.New("Destination must by struct or a pointer to struct")
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
// Keep in mind that random seed must be initialized before. Example:
// 		rand.Seed(time.Now().Unix())
func Random(min, max int) int {
	return rand.Intn(max-min) + min
}

// RandomString generates a random string of the specified length.
// Keep in mind that random seed must be initialized before. Example:
// 		rand.Seed(time.Now().Unix())
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterAndNumberRunes[rand.Intn(len(letterAndNumberRunes))]
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
