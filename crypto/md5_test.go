package crypto

import "testing"

func TestMD5(t *testing.T) {
	str := "text for testing"
	hashedOK := "35de525c58bd667291e9d7826542e500"
	hashedTest := GetMD5Hash(str)

	if hashedTest != hashedOK {
		t.Errorf("Invaild hash: %s != %s", hashedTest, hashedOK)
	}
}
