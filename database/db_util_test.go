package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	ID       int     `db:"id_user"`
	Name     *string `db:"name" db_type:"varchar"`
	Email    *string `db:"email" db_type:"varchar"`
	Address  *string `db:"address" db_type:"varchar"`
	Password *string `db:"password" db_type:"varchar"`
	City     *string
	Country  *string `db_type:"varchar"`
	Active   bool    `db_type:"boolean"`
}

// test cases for BuildUpdateSetQuery()
func TestBuildUpdateSetQuery(t *testing.T) {

	witnesUserStr := "SET `password`='myhashedpassword',`address`='Luis Bonavita 1122',`id_user`=145,`city`='Montevideo',`country`='UY',`active`=true"

	name := "Pepe"
	email := "pepe@astropay.com"
	password := "myhashedpassword"
	address := "Luis Bonavita 1122"
	city := "Montevideo"
	country := "UY"

	// test obj
	user := &User{
		ID:       145,
		Name:     &name,
		Email:    &email,
		Password: &password,
		Address:  &address,
		City:     &city,
		Country:  &country,
		Active:   true,
	}

	// fields to update
	dirtyFields := []string{"Password", "Address", "ID", "City", "Country", "Active"}
	builtSetStr, err := BuildUpdateSetQuery(user, dirtyFields)
	if err == nil {
		if builtSetStr != witnesUserStr {
			t.Errorf("BuildUpdateSetQuery() returned a wrong string: %s", builtSetStr)
		}
	} else {
		t.Errorf("BuildUpdateSetQuery() returned an error: %s", err.Error())
	}

	// fields to update with a wrong field
	dirtyFieldsInvalid := []string{"Password", "Address", "ID", "Status"}
	_, err = BuildUpdateSetQuery(user, dirtyFieldsInvalid)
	if err == nil {
		t.Errorf("BuildUpdateSetQuery() should have returned an error")
	}
}

// test cases for BuildParametrizedUpdateSetQuery()
func TestBuildParametrizedUpdateSetQuery(t *testing.T) {

	witnessUserStr := "SET password=?,id_user=?,country=?,active=?"

	name := "Pepe"
	email := "pepe@astropay.com"
	password := "myhashedpassword"
	country := "UY"

	// test obj
	user := User{
		ID:       145,
		Name:     &name,
		Email:    &email,
		Password: &password,
		Country:  &country,
		Active:   true,
	}

	// fields to update
	dirtyFields := []string{"Password", "ID", "Country", "Active"}
	builtSetStr, err := BuildParametrizedUpdateSetQuery(user, dirtyFields)
	if err == nil {
		if builtSetStr != witnessUserStr {
			t.Errorf("BuildParametrizedUpdateSetQuery() returned a wrong string: %s", builtSetStr)
		}
	} else {
		t.Errorf("BuildParametrizedUpdateSetQuery() returned an error: %s", err.Error())
	}

	// fields to update with a wrong field
	dirtyFieldsInvalid := []string{"Password", "Address", "ID", "Status"}
	_, err = BuildParametrizedUpdateSetQuery(user, dirtyFieldsInvalid)
	if err == nil {
		t.Errorf("BuildParametrizedUpdateSetQuery() should have returned an error")
	}
}

// test cases for BuildNamedParametersUpdateSetQuery()
func TestBuildNamedParamersUpdateSetQuery(t *testing.T) {

	witnessUserStr := "SET password=:password,id_user=:id_user,country=:country,active=:active"

	name := "Pepe"
	email := "pepe@astropay.com"
	password := "myhashedpassword"
	country := "UY"

	// test obj
	user := User{
		ID:       145,
		Name:     &name,
		Email:    &email,
		Password: &password,
		Country:  &country,
		Active:   true,
	}

	// fields to update
	dirtyFields := []string{"Password", "ID", "Country", "Active"}
	builtSetStr, err := BuildNamedParametersUpdateSetQuery(user, dirtyFields)
	if err == nil {
		if builtSetStr != witnessUserStr {
			t.Errorf("BuildNamedParametersUpdateSetQuery() returned a wrong string: %s", builtSetStr)
		}
	} else {
		t.Errorf("BuildNamedParametersUpdateSetQuery() returned an error: %s", err.Error())
	}

	// fields to update with a wrong field
	dirtyFieldsInvalid := []string{"Password", "Address", "ID", "Status"}
	_, err = BuildNamedParametersUpdateSetQuery(user, dirtyFieldsInvalid)
	if err == nil {
		t.Errorf("BuildNamedParametersUpdateSetQuery() should have returned an error")
	}
}

// test cases for GetParameterValues()
func TestGetParameterValues(t *testing.T) {

	id := 100
	name := "Pepe"
	email := "pepe@astropay.com"
	password := "myhashedpassword"
	country := "UY"
	active := true

	// test obj
	user := &User{
		ID:       id,
		Name:     &name,
		Email:    &email,
		Password: &password,
		Country:  &country,
		Active:   active,
	}

	dirtyFields := []string{"Password", "ID", "Country", "Active"}
	params, err := GetParameterValues(user, dirtyFields)
	if err != nil {
		t.Errorf("GetParameterValues() returned error: %s", err.Error())
	} else {
		assert.Equal(t, password, params[0])
		assert.Equal(t, id, params[1])
		assert.Equal(t, country, params[2])
		assert.Equal(t, active, params[3])
	}

	// fields to update with a wrong field
	dirtyFieldsInvalid := []string{"Password", "ID", "Status"}
	_, err = GetParameterValues(user, dirtyFieldsInvalid)
	if err == nil {
		t.Errorf("GetParameterValues() should have returned an error")
	}
}

// test cases for GetAllFields()
func TestGetAllFields(t *testing.T) {

	witnessStr := "id_user,name,email,address,password,city,country,active"
	witnessStrQuoted := "`id_user`,`name`,`email`,`address`,`password`,`city`,`country`,`active`"
	witnessStrNamedParam := ":id_user,:name,:email,:address,:password,:city,:country,:active"

	// test obj
	user := new(User)

	// normal
	fieldList, err := GetAllFields(user, false, false)
	if err != nil {
		t.Error(err.Error())
	} else {
		if fieldList != witnessStr {
			t.Logf("expected: %s, Got: %s", witnessStr, fieldList)
		}
	}

	// quoted
	fieldListQuoted, errQuoted := GetAllFields(user, true, false)
	if errQuoted != nil {
		t.Error(errQuoted.Error())
	} else {
		if fieldListQuoted != witnessStrQuoted {
			t.Logf("expected: %s, Got: %s", witnessStrQuoted, fieldListQuoted)
		}
	}

	// named parameters
	fieldListNamedParam, errNamed := GetAllFields(user, true, false)
	if errNamed != nil {
		t.Error(errNamed.Error())
	} else {
		if fieldListNamedParam != witnessStrNamedParam {
			t.Logf("expected: %s, Got: %s", witnessStrNamedParam, fieldListNamedParam)
		}
	}
}
