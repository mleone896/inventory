package models

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	dbp "github.com/mleone896/inventory/db"
)

// Color representation of a color
type Color struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	InUse     bool      `json:"in_use"`
	LastInUse time.Time `json:"last_in_use"`
}

// ColoConfigFun type allows function option configuration
type ColoConfigFun func(*Color)

// SQLGetUnusedColors will never change so save memory by constant
const SQLGetUnusedColors = `
    SELECT * FROM colors
	WHERE in_use = false
	AND
	last_in_use < (NOW() - INTERVAL '24 hours')
    LIMIT 100
	`

// NewColor constructor for a new color
func NewColor(opts ...func(*Color)) (*Color, error) {
	col := &Color{}

	// make sure we were passed something
	if len(opts) >= 1 {
		for _, opt := range opts {
			opt(col)
		}
	}

	return col, nil
}

// WithDefault initializing blank color
func WithDefault() ColoConfigFun {
	return func(c *Color) {
	}
}

// WithName sets the name of color called when initializing
func WithName(name string) ColoConfigFun {
	return func(c *Color) {
		c.Name = name
	}
}

// WithID sets the ID of color called when initializing
func WithID(id int) ColoConfigFun {
	return func(c *Color) {
		c.ID = id
	}
}

// Get will retrieve the next color to be used mark it as used update the timestamp
// and set the appropriate fields in the struct
func (c *Color) Get(db *sqlx.DB) error {
	colors := []Color{}
	tx := db.MustBegin()
	defer tx.Rollback()

	if err := tx.Select(&colors, SQLGetUnusedColors); err != nil {
		return err
	}
	colorNames := make([]string, len(colors))
	for i, col := range colors {
		colorNames[i] = col.Name

	}

	randIndex := rand.Intn(len(colorNames))
	picked := colorNames[randIndex]
	tx.MustExec(`UPDATE colors SET in_use = true, last_in_use = NOW() WHERE name = $1`, picked)

	err := tx.Commit()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("could not commit transaction: %v", err)
	}
	err = db.QueryRowx("SELECT * from colors where name = $1", picked).StructScan(c)
	if err != nil {
		return err
	}

	return nil
}

// Update marks a color as being used
func (c *Color) Update(db *sqlx.DB) error {
	tx := db.MustBegin()

	defer tx.Rollback()
	tx.MustExec(`UPDATE colors SET in_use = true, last_in_use = NOW() WHERE name = $1`, c.Name)

	err := tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Colors ...
func Colors() *Color {
	co, _ := NewColor(WithDefault())

	return co
}

// ColorSlice ...
type ColorSlice []*Color

// FindAll ...
func (c *Color) FindAll(db *sqlx.DB) (*sqlx.Rows, error) {

	rows, err := db.Queryx("SELECT * from colors where in_use = 'f'")

	if err != nil {
		return nil, fmt.Errorf("could not select from color: %s", err)
	}
	return rows, nil

}

// UnpackRows takes a sql.Rows and scans into a struct slice
func (c Color) UnpackRows(rows *sqlx.Rows) ([]Color, error) {
	cs := []Color{}

	if err := sqlx.StructScan(rows, &cs); err != nil {
		return nil, fmt.Errorf("could not scan rows into slice %s", err)
	}

	return cs, nil

}

// Sync marks a color as being used
func (c Color) Sync(db *sqlx.DB, colors []string) error {

	// now we start a tx, update in_use to false and set in_use to true
	// with the returned colors from aws... aws is the source of truth
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("could not get tx handler: %s", err)
	}

	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE colors SET in_use = false`)
	if err != nil {
		return dbp.TxRollbackHandleError(tx, err)
	}

	// Loop through all the colors and set the ones in use to true
	for _, color := range colors {
		_, err := tx.Exec(`UPDATE colors SET in_use = true, last_in_use = NOW() WHERE name = $1`, color)

		if err != nil {
			return dbp.TxRollbackHandleError(tx, err)
		}

	}

	return dbp.TxCommitHandleError(tx)
}

// Delete ...
func (c *Color) Delete(db *sqlx.DB) error {

	tx, err := db.Beginx()

	if err != nil {
		return fmt.Errorf("error could not begin got: %s", err)
	}
	defer func() {
		if rerr := tx.Rollback(); rerr != nil && rerr != sql.ErrTxDone {
			log.Printf("failed to rollback subnet tx: %+v", rerr)
		}
	}()

	_, err = tx.Exec("DELETE FROM colors")
	if err != nil {
		return dbp.TxRollbackHandleError(tx, err)
	}

	return dbp.TxCommitHandleError(tx)

}
