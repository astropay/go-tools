package files

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileExists(t *testing.T) {
	// create temp file and check that exists
	file, err := ioutil.TempFile(os.TempDir(), "test")
	if err != nil {
		t.Errorf("Error creating temp file '%s': %s", file.Name(), err.Error())
	}
	defer os.Remove(file.Name())

	if !Exists(file.Name()) {
		t.Errorf("File '%s' is not present", file.Name())
	}

	// fake path
	if Exists("/not/a/real/path") {
		t.Error("Test failed checking a fake path")
	}
}
