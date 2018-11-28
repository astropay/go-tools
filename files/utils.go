/*
	@author Robert
*/

package files

import (
	"errors"
	"os"
	"path/filepath"
)

// File errors
var (
	ErrGetFileInfo       = errors.New("Error get FileInfo")
	ErrFileAlreadyExists = errors.New("File already exists")
	ErrCreateFile        = errors.New("Error create file")
)

// Exists returns true/false depending if the file indicated in path, exists or not
func Exists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

// CreateFile create a file from path.
func CreateFile(path string) (err error) {
	_, err = os.Stat(path)

	if os.IsExist(err) {
		return ErrFileAlreadyExists
	}

	var file *os.File
	file, err = os.Create(path)
	defer file.Close()
	if err != nil {
		return ErrCreateFile
	}

	return
}

// DeleteFile delete a file from path
func DeleteFile(path string) (err error) {
	// delete file
	err = os.Remove(path)
	return
}

// GetAppPath returns the application absolute path
func GetAppPath() string {
	if dir, err := filepath.Abs(filepath.Dir(os.Args[0])); err == nil {
		return dir
	}

	return ""
}
