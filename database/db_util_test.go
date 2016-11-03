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

	witnesUserStr := "SET password=?,id_user=?,country=?,active=?"

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
		if builtSetStr != witnesUserStr {
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
