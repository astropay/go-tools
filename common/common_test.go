package common

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// test for valid email address
func TestIsEmailAddress(t *testing.T) {
	almostEmail := "user@host"
	if valid := IsEmailAddress(almostEmail); valid {
		t.Errorf("'%s' is not a valid address", almostEmail)
	}

	isEmail := "user@host.com"
	if valid := IsEmailAddress(isEmail); !valid {
		t.Errorf("'%s' is a valid email address", isEmail)
	}

	notEmail := "just a normal string"
	if valid := IsEmailAddress(notEmail); valid {
		t.Errorf("'%s' is not a valid address", notEmail)
	}
}

// test for UpdateStructFields() function
func TestUpdateStructFields(t *testing.T) {

	type User struct {
		ID    int
		Name  string
		Email string
	}

	// test OK 1
	user1 := &User{ID: 1, Name: "User 1"}
	assert.Equal(t, "", user1.Email)

	newData1 := make(map[string]interface{})
	newData1["Email"] = "user1@astropay.com"

	if err := UpdateStructFromMap(user1, newData1); err != nil {
		t.Errorf("UpdateStructFields() failed: %s", err.Error())
	} else {
		assert.Equal(t, 1, user1.ID)
		assert.Equal(t, "User 1", user1.Name)
		assert.Equal(t, "user1@astropay.com", user1.Email)
	}

	// test OK 2
	user2 := &User{ID: 2, Name: "User 2", Email: "user2@astropay.com"}
	assert.Equal(t, "user2@astropay.com", user2.Email)

	newData2 := make(map[string]interface{})
	newData2["Name"] = "User Two"
	newData2["Email"] = ""
	newData2["UnkownField"] = "This value won't be set anywhere"
	newData2["UnkownField2"] = false

	if err := UpdateStructFromMap(user2, newData2); err != nil {
		t.Errorf("UpdateStructFields() failed: %s", err.Error())
	} else {
		assert.Equal(t, 2, user2.ID)
		assert.Equal(t, "User Two", user2.Name)
		assert.Equal(t, "", user2.Email)
	}

	// test err
	notAStruct := "Not really a struct, just a string"
	newData3 := make(map[string]interface{})
	newData3["Field"] = "Field value"

	if err := UpdateStructFromMap(notAStruct, newData3); err == nil {
		t.Errorf("UpdateStructFields() should have failed with ErrNotStruct")
	}

}

func TestRandom(t *testing.T) {
	rand.Seed(time.Now().Unix())

	if test := Random(1, 10); test < 1 || test > 10 {
		t.Errorf("Random number should be between 1 and 10: %v", test)
	}

	if test := Random(1, 100); test < 1 || test > 100 {
		t.Errorf("Random number should be between 1 and 100: %v", test)
	}

	if test := Random(12345, 989912); test < 12345 || test > 989912 {
		t.Errorf("Random number should be between 12345 and 989912: %v", test)
	}

}
