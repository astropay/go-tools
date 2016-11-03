/*
	@author Robert
*/

package files

import (
	"os"
)

// Exists returns true/false depending if the file indicated in path, exists or not
func Exists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
