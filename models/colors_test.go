package models

import (
	"fmt"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func returnColorCols() []string {
	cols := []string{"id", "name", "in_use", "last_in_use"}
	return cols
}

func returnSingleColor() *sqlmock.Rows {
	rows := sqlmock.NewRows(returnColorCols()).
		AddRow(1, "orange", true, time.Now())
	return rows
}

func returnManyColors() *sqlmock.Rows {
	rows := sqlmock.NewRows(returnColorCols()).
		AddRow(1, "orange", true, time.Now()).
		AddRow(2, "green", true, time.Now()).
		AddRow(3, "blue", true, time.Now()).
		AddRow(4, "red", true, time.Now()).
		AddRow(5, "purple", true, time.Now()).
		AddRow(6, "rubyred", false, time.Now()).
		AddRow(7, "yellow", true, time.Now())

	return rows
}

func TestNewColor(t *testing.T) {
	color, err := NewColor(WithDefault())
	errCheck(err, t)

	// use reflection to figure out if the default config was initiazlized correctly
	if !color.IsEmpty() {
		t.Fatal(fmt.Errorf("newcolor expected default vals to be empty config got %v", color))
	}
}

func TestNewColorWithName(t *testing.T) {
	color, err := NewColor(WithName("orange"))
	errCheck(err, t)
	if color.Name != "orange" {
		t.Fatal(fmt.Errorf("newcolor expected name to orange config got %v", color.Name))
	}

}

func TestNewColorWithID(t *testing.T) {
	color, err := NewColor(WithID(1))
	errCheck(err, t)
	if color.ID != 1 {
		t.Fatal(fmt.Errorf("newcolor expected id to 1 config got %v", color.ID))
	}

}

func TestGet(t *testing.T) {
	mod, mock := initTestDB()
	defer mod.Conn.Close()

	expectColor := "orange"

	mock.ExpectBegin()
	mock.ExpectQuery(".*").
		WillReturnRows(returnSingleColor())

	mock.ExpectExec("UPDATE colors SET .*").
		WithArgs("orange").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()
	srow := sqlmock.NewRows(returnColorCols()).
		AddRow(1, expectColor, true, time.Now())

	mock.ExpectQuery("SELECT .*").
		WithArgs(expectColor).
		WillReturnRows(srow)

	color, err := NewColor(WithDefault())
	if err != nil {
		t.Error(err)
	}

	err = color.Get(mod.Conn)

	if err != nil {
		t.Fatalf("expected error to return nil got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there are unfulfilled expectations: %s", err)
	}

	if color.Name != expectColor {
		t.Fatalf("wrong color set expected %s, got: %s", expectColor, color.Name)
	}

	if !color.InUse {
		t.Fatalf("expected in_use to be updated to true got false")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there are unfulfilled expectations: %s", err)
	}

}

func TestUpdate(t *testing.T) {
	mod, mock := initTestDB()

	defer mod.Conn.Close()
	expectColor := "orange"
	c, err := NewColor(WithName(expectColor))
	errCheck(err, t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE colors SET in_use = true, last_in_use = .*").
		WithArgs(expectColor).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	if err := c.Update(mod.Conn); err != nil {
		t.Fatalf("expected color to update got %v", err)
	}

}

func TestGetAll(t *testing.T) {
	mod, mock := initTestDB()
	defer mod.Conn.Close()

	mock.ExpectQuery(".*").
		WillReturnRows(returnManyColors())

	rows, err := mod.FindAll(Colors())
	errCheck(err, t)

	cs := []*Color{}
	if err = sqlx.StructScan(rows, &cs); err != nil {
		t.Errorf("could not scan rows into slice %s", err.Error())
	}

	if err != nil {
		t.Fatalf("expected color to update got %v", err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there are unfulfilled expectations: %s", err)
	}

}
