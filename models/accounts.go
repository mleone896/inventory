package models

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	dbp "github.com/mleone896/inventory/db"
)

// Account represents the accuont to make operations on
type Account struct{ id string }

// NewAccount returns an account object
func NewAccount(id string) *Account {
	return &Account{
		id: id,
	}
}

// Get returns a subnet from accountid
func (a *Account) Get(db *sqlx.DB) error {
	return nil
}

// Sync returns a func that updates all subnets
func (a *Account) Sync(db *sqlx.DB, subs []*Subnet) error {

	if err := validateAccountID(subs, a.id); err != nil {
		return err
	}

	// create the first class func to insert the subnets by account
	insert := `
		INSERT INTO subnets (
			vpc_id,
			subnet_id,
			availability_zone,
			account_id,
			tags
		)
		VALUES ( 
		:vpc_id, 
		:subnet_id, 
		:availability_zone, 
		:account_id, 
		:tags)`

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("error beginning transaction %v", err)
	}

	defer func() {
		if rerr := tx.Rollback(); rerr != nil && rerr != sql.ErrTxDone {
			log.Printf("failed to rollback subnet tx: %+v", rerr)
		}
	}()

	_, err = tx.Exec(`TRUNCATE table subnets restart identity`)

	if err != nil {
		return dbp.TxRollbackHandleError(tx, err)
	}

	stmt, err := tx.PrepareNamed(insert)
	if err != nil {
		return fmt.Errorf("error preparing subnet stmt: %v", err)
	}

	// loop over the subnets and create the exec statement
	for _, sub := range subs {
		_, err := stmt.Exec(sub)

		// if there is an error check tx.Rollback for a error and return
		// original
		if err != nil {
			return dbp.TxRollbackHandleError(tx, err)
		}
	}

	if err := stmt.Close(); err != nil {
		return fmt.Errorf("error triggerd when closing stmt for subnets %v", err)
	}

	return dbp.TxCommitHandleError(tx)

}

// helper function to ensure all items in subnet slice have the correct id
func validateAccountID(subs []*Subnet, accountID string) error {
	// loop through all the items in the slice and confirm they have the appropriate
	// account id
	for _, sub := range subs {
		if sub.AccountID != accountID {
			return fmt.Errorf("error expected account id %s got: %s", accountID, sub.AccountID)
		}
	}

	return nil
}
