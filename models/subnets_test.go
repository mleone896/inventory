package models

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestSubnetDeletion(t *testing.T) {
	mod, mock := initTestDB()
	defer mod.Conn.Close()
	sub := Subnets()
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM .*").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	if err := mod.Delete(sub); err != nil {
		t.Errorf("expected no error got: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}

}
