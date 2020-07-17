package models

import (
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func returnInstanceCols() []string {
	cols := []string{"id", "instance_id", "account_id", "tags"}
	return cols
}

// returns a sqlmock rows with many Instance
func returnManyInstancesRows() *sqlmock.Rows {

	rows := sqlmock.NewRows(returnInstanceCols())
	instances := returnManyInstances()
	for _, inst := range instances {
		rows.AddRow(inst.ID, inst.InstanceID, inst.AccountID, inst.Tags)
	}

	return rows
}

func returnSingleInstance() *Instance {
	return &Instance{
		InstanceID: "i-0161c8cb6bfdea7f3",
		AccountID:  "238967563593",
		SubnetID:   "subnet-295fcf02",
		Tags:       hstoreHelper("foo"),
	}
}

func returnManyInstances() []*Instance {
	Instances := []*Instance{
		&Instance{
			InstanceID: "i-0161c8cb6bfdea7f3",
			AccountID:  "238967563593",
			SubnetID:   "subnet-295fcf02",
			Tags:       hstoreHelper("foo"),
		},
		&Instance{
			InstanceID: "i-03c6f3b2f73a120bc",
			AccountID:  "238967563593",
			SubnetID:   "subnet-295fcf02",
			Tags:       hstoreHelper("bar"),
		},
		&Instance{
			InstanceID: "i-07af2cf863c58a6d0",
			AccountID:  "238967563593",
			SubnetID:   "subnet-295fcf02",
			Tags:       hstoreHelper("baz"),
		},
		&Instance{
			InstanceID: "i-10816c8f",
			AccountID:  "238967563593",
			SubnetID:   "subnet-295fcf02",
			Tags:       hstoreHelper("bing"),
		},
		&Instance{
			InstanceID: "i-11816c8e",
			AccountID:  "238967563593",
			SubnetID:   "subnet-295fcf02",
			Tags:       hstoreHelper("thing"),
		},
		&Instance{
			InstanceID: "i-1f816c80",
			AccountID:  "238967563593",
			SubnetID:   "subnet-295fcf02",
			Tags:       hstoreHelper("enigma"),
		},
		&Instance{
			InstanceID: "i-20afcc88",
			AccountID:  "238967563593",
			SubnetID:   "subnet-295fcf02",
			Tags:       hstoreHelper("namer"),
		},
	}

	return Instances
}

func TestInstanceSync(t *testing.T) {
	mod, mock := initTestDB()
	defer mod.Conn.Close()

	// test all valid entries
	a := NewInstance(WithDefaultInstance())

	mul := returnManyInstances()
	mock.ExpectBegin()
	prepare := mock.ExpectPrepare(".*")
	for _, inst := range mul {
		prepare.ExpectExec().
			WithArgs(
				inst.InstanceID,
				inst.AccountID,
				inst.SubnetID,
				inst.Tags,
				inst.Tags,
				inst.InstanceID,
				inst.AccountID).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	mock.ExpectCommit()

	err := a.Sync(mod.Conn, mul)

	errCheck(err, t)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %+v", err)
	}

}

func TestInstanceSyncRollback(t *testing.T) {
	mod, mock := initTestDB()
	defer mod.Conn.Close()

	// test all valid entries
	a := NewInstance(WithDefaultInstance())

	instance := returnSingleInstance()

	mock.ExpectBegin()
	mock.ExpectPrepare(".*").
		ExpectExec().
		WithArgs(
			instance.InstanceID,
			instance.AccountID,
			instance.SubnetID,
			instance.Tags,
			instance.Tags,
			instance.InstanceID,
			instance.AccountID).
		WillReturnError(fmt.Errorf("some testing error instance"))

	mock.ExpectRollback()

	mul := []*Instance{instance}

	err := a.Sync(mod.Conn, mul)

	if err == nil {
		t.Errorf("expected err but got: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %+v", err)
	}

}
