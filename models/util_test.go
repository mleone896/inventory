package models

import (
	"log"
	"reflect"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/mleone896/inventory/db"
)

// helper function
func (c Color) IsEmpty() bool {
	return reflect.DeepEqual(c, Color{})
}

func errCheck(err error, t *testing.T) {
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func initTestDB() (*db.DataObj, sqlmock.Sqlmock) {

	//	caller must close the db connection
	dbm, mock, err := sqlmock.New()

	if err != nil {
		log.Fatalf("an error %s was not expected when openeing a stub database connection", err)
	}

	sqlxDB := sqlx.NewDb(dbm, "sqlmock")

	obj := &db.DataObj{Conn: sqlxDB}
	obj.Conn.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
	return obj, mock
}
