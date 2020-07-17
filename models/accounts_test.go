package models

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq/hstore"
)

func returnSubnetsCols() []string {
	cols := []string{"id",
		"subnet_id",
		"account_id",
		"availability_zone",
		"tags",
		"vpc_id",
	}
	return cols
}

func hstoreHelper(s string) hstore.Hstore {
	data := make(map[string]sql.NullString, 1)
	data["Name"] = sql.NullString{String: s, Valid: true}
	hs := hstore.Hstore{Map: data}
	return hs
}

func returnSubnetCols() []string {
	cols := []string{"id",
		"subnet_id",
		"account_id",
		"availability_zone",
		"tags",
		"vpc_id"}
	return cols
}

// returns a sqlmock rows with 1 subnet
func returnSingleSubnetRow() *sqlmock.Rows {
	rows := sqlmock.NewRows(returnSubnetCols()).
		AddRow(1, "orange", true, time.Now())
	return rows
}

// adds a wrong record to sqlmock rows NOTE(not pure mutates)
func subWithWrongData(rws *sqlmock.Rows, sub *Subnet) *sqlmock.Rows {

	return rws.AddRow(sub.ID,
		sub.SubnetID,
		sub.AccountID,
		sub.AZ,
		sub.Tags,
		sub.VpcID,
	)
}

// returns a sqlmock rows with many subnet
func returnManySubnetsRows() *sqlmock.Rows {

	rows := sqlmock.NewRows(returnSubnetCols()).
		AddRow(1, "subnet-295fcf02", "238967563593", "us-east-1c", hstoreHelper("foo"), "vpc-df4a70ba").
		AddRow(2, "subnet-20eaa40b", "238967563593", "us-east-1c", hstoreHelper("bar"), "vpc-df4a70ba").
		AddRow(3, "subnet-1422bd3e", "238967563593", "us-east-1c", hstoreHelper("baz"), "vpc-df4a70ba").
		AddRow(4, "subnet-4b5fcf60", "238967563593", "us-east-1c", hstoreHelper("bing"), "vpc-df4a70ba").
		AddRow(5, "subnet-3d8be516", "238967563593", "us-east-1c", hstoreHelper("thing"), "vpc-df4a70ba").
		AddRow(6, "subnet-c16010ea", "238967563593", "us-east-1c", hstoreHelper("enigma"), "vpc-df4a70ba").
		AddRow(7, "subnet-325fcf19", "238967563593", "us-east-1c", hstoreHelper("namer"), "vpc-df4a70ba")

	return rows
}

func returnSingleSubnet() *Subnet {
	return &Subnet{
		SubnetID:  "subnet-295fcf02",
		AccountID: "238967563593",
		AZ:        "us-east-1c",
		Tags:      hstoreHelper("foo"),
		VpcID:     "vpc-df4a70ba",
	}
}

func returnManySubnets() []*Subnet {
	subnets := []*Subnet{
		&Subnet{
			SubnetID:  "subnet-295fcf02",
			AccountID: "238967563593",
			AZ:        "us-east-1c",
			Tags:      hstoreHelper("foo"),
			VpcID:     "vpc-df4a70ba",
		},
		&Subnet{
			SubnetID:  "subnet-20eaa40b",
			AccountID: "238967563593",
			AZ:        "us-east-1c",
			Tags:      hstoreHelper("bar"),
			VpcID:     "vpc-df4a70ba",
		},
		&Subnet{
			SubnetID:  "subnet-1422bd3e",
			AccountID: "238967563593",
			AZ:        "us-east-1c",
			Tags:      hstoreHelper("baz"),
			VpcID:     "vpc-df4a70ba",
		},
		&Subnet{
			SubnetID:  "subnet-4b5fcf60",
			AccountID: "238967563593",
			AZ:        "us-east-1c",
			Tags:      hstoreHelper("bing"),
			VpcID:     "vpc-df4a70ba",
		},
		&Subnet{
			SubnetID:  "subnet-3d8be516",
			AccountID: "238967563593",
			AZ:        "us-east-1c",
			Tags:      hstoreHelper("thing"),
			VpcID:     "vpc-df4a70ba",
		},
		&Subnet{
			SubnetID:  "subnet-3d8be516",
			AccountID: "238967563593",
			AZ:        "us-east-1c",
			Tags:      hstoreHelper("enigma"),
			VpcID:     "vpc-df4a70ba",
		},
		&Subnet{
			SubnetID:  "subnet-325fcf19",
			AccountID: "238967563593",
			AZ:        "us-east-1c",
			Tags:      hstoreHelper("namer"),
			VpcID:     "vpc-df4a70ba",
		},
	}

	return subnets
}

func TestAccountSync(t *testing.T) {

	mod, mock := initTestDB()

	defer mod.Conn.Close()

	// test all valid entries
	a := NewAccount("238967563593")

	subnets := returnManySubnets()

	mock.ExpectBegin()
	// make sure it deletes and restarts pk
	mock.ExpectExec("TRUNCATE.*").
		WillReturnResult(sqlmock.NewResult(0, 0))

	prepare := mock.ExpectPrepare(".*")

	var count int64
	count = 1
	for _, sub := range subnets {
		prepare.ExpectExec().
			WithArgs(
				sub.VpcID,
				sub.SubnetID,
				sub.AZ,
				sub.AccountID,
				sub.Tags).
			WillReturnResult(sqlmock.NewResult(1, count))
		count++
	}

	mock.ExpectCommit()

	err := a.Sync(mod.Conn, subnets)

	errCheck(err, t)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Was expecting no error got %v", err)
	}

	item := &Subnet{
		SubnetID:  "subnet-325fcf19",
		AccountID: "438967563593",
		AZ:        "us-east-1c",
		Tags:      hstoreHelper("namer"),
		VpcID:     "vpc-df4a70ba",
	}

	subnets = append(subnets, item)
	err = a.Sync(mod.Conn, subnets)

	if err == nil {
		t.Errorf("expected error got: %v", err)
	}

}

func TestAccountSyncRollback(t *testing.T) {
	mod, mock := initTestDB()
	defer mod.Conn.Close()

	// test all valid entries
	a := NewAccount("238967563593")

	sub := returnSingleSubnet()

	mul := []*Subnet{sub}

	mock.ExpectBegin()
	// make sure it deletes
	mock.ExpectExec("TRUNCATE table subnets*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	//
	mock.ExpectPrepare(".*").
		ExpectExec().
		WithArgs(
			sub.VpcID,
			sub.SubnetID,
			sub.AZ,
			sub.AccountID,
			sub.Tags).
		WillReturnError(fmt.Errorf("Testing Rollback Error"))

	mock.ExpectRollback()

	err := a.Sync(mod.Conn, mul)

	if err == nil {
		t.Error("expected error but got none", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}
