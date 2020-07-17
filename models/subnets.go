package models

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq/hstore"
)

// Subnet table
type Subnet struct {
	ID        int           `json:"id"`
	SubnetID  string        `json:"subnet_id" db:"subnet_id"`
	AccountID string        `json:"account_id" db:"account_id"`
	AZ        string        `json:"availability_zone" db:"availability_zone"`
	Tags      hstore.Hstore `json:"tags" db:"tags"`
	VpcID     string        `json:"vpc_id" db:"vpc_id"`
}

// SubnetConfigFun type allows function option configuration
type SubnetConfigFun func(*Subnet)

// NewSubnet constructor for a new color
func NewSubnet(opts ...func(*Subnet)) (*Subnet, error) {
	sub := &Subnet{}

	// make sure we were passed something
	if len(opts) >= 1 {
		for _, opt := range opts {
			opt(sub)
		}
	}

	return sub, nil
}

// WithDefaultSubnet initializing blank subnet
func WithDefaultSubnet() SubnetConfigFun {
	return func(c *Subnet) {
	}
}

// WithSubnetID sets the subnetid when initializing
func WithSubnetID(id string) SubnetConfigFun {
	return func(c *Subnet) {
		c.SubnetID = id
	}
}

// Create inserts a subnet into the database
func (s *Subnet) Create(db *sqlx.DB) error {

	return nil
}

// Get statisfies the Getter interface this is really get by subnet_id and not
// the id on the subnet table
func (s *Subnet) Get(db *sqlx.DB) error {

	err := db.QueryRowx("SELECT * from subnets where subnet_id = $1", s.SubnetID).StructScan(s)

	if err != nil {
		return fmt.Errorf("could not find subnet %v", err)
	}

	return nil

}

// Subnets ...
func Subnets() *Subnet {
	s, _ := NewSubnet(WithDefaultSubnet())
	return s
}

// Delete ...
func (s Subnet) Delete(db *sqlx.DB) error {

	tx, err := db.Beginx()

	if err != nil {
		return fmt.Errorf("error could not begin got: %s", err)
	}
	defer func() {
		if rerr := tx.Rollback(); rerr != nil && rerr != sql.ErrTxDone {
			log.Printf("failed to rollback subnet tx: %+v", rerr)
		}
	}()

	_, err = tx.Exec("DELETE FROM subnets")
	if err != nil {
		terr := tx.Rollback()
		if terr != nil {
			return fmt.Errorf("tx caused error: %s", terr)
		}
		return fmt.Errorf("error could not exec delete: %s", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error could not commit transaction %s", err)
	}

	return nil

}
