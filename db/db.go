package db

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"

	// linter:
	_ "github.com/lib/pq"
)

// ConfigFunc allows caller to set config options without accessing the un
// exported structs fields directly
type ConfigFunc func(*DataObj)

// ExecFun is a representation of a function that only needs to take a connection
type ExecFun func(*sqlx.DB) error

// Getter interface for anything that knows how to get
type Getter interface {
	Get(*sqlx.DB) error
}

// Updater interface for objects that know how to update
type Updater interface {
	Update(*sqlx.DB) error
}

type Reader interface {
	Read(Getter) error
}

// Deleter empty a table
type Deleter interface {
	Delete(*sqlx.DB) error
}

// Inserter takes anything that can insert a record
type Inserter interface {
	Create(*sqlx.DB) (string, error)
}

// SelectAller needs a better name
type SelectAller interface {
	FindAll(*sqlx.DB) (*sqlx.Rows, error)
}

// DBer holds all interfaces
type DBer interface {
	Getter
	Updater
	SelectAller
}

// DataObj connection info object
type DataObj struct {
	connString string
	Conn       *sqlx.DB
}

// needed to pick the random color
func init() {
	t := time.Now()
	rand.Seed(int64(t.Nanosecond()))
}

// New returns a new Data access object
func New(options ...func(*DataObj)) (*DataObj, error) {

	dao := &DataObj{connString: "dbname=inventory sslmode=disable"}
	if len(options) >= 1 {

		for _, option := range options {
			option(dao)
		}

	}

	c, err := sqlx.Open("postgres", dao.connString)
	if err != nil {
		return nil, fmt.Errorf("could not open database connection: %v", err)
	}

	dao.Conn = c

	// mlcrsi: instead of having db tags and json tags just use json
	dao.Conn.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)

	return dao, nil
}

// WithConnString sets the postgres connection string
func WithConnString(s string) ConfigFunc {
	return func(d *DataObj) {
		d.connString = s
	}
}

// TxRollbackHandleError takes a sqlx Transaction and tries to roll it back
// if that doesn't succeed it returns that error, if it does it returns the
// error that was produced from the Exec
func TxRollbackHandleError(tx *sqlx.Tx, execError error) error {
	err := tx.Rollback()
	if err != nil {
		return fmt.Errorf("txrollbackerror: could not rollback transaction:  erroris %s erroris", err)
	}
	return execError
}

// TxCommitHandleError takes a transaction and tries to commit if it fails returns
// the error else returns nil
func TxCommitHandleError(tx *sqlx.Tx) error {
	err := tx.Commit()
	if err != nil {
		return fmt.Errorf("txcommithandleerror: could not commit transaction:  %s", err)
	}

	return nil
}
