package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Find takes anything that knows how to get and propogrates the error up the
// call stack
func (d *DataObj) Read(q Getter) error {

	err := q.Get(d.Conn)
	if err != nil {
		return fmt.Errorf("could not find from queryer: %v", err)
	}

	return err
}

// FindAll ...
func (d *DataObj) FindAll(f SelectAller) (*sqlx.Rows, error) {
	return f.FindAll(d.Conn)
}

// Create ...
func (d *DataObj) Create(c Inserter) (string, error) {
	// create the thing and capture the id
	id, err := c.Create(d.Conn)
	if err != nil {
		return "", err
	}
	return id, nil
}

// Update ....
func (d *DataObj) Update(u Updater) error {
	if err := u.Update(d.Conn); err != nil {
		return err
	}

	return nil
}

// Delete ...
func (d *DataObj) Delete(de Deleter) error {

	if err := de.Delete(d.Conn); err != nil {
		return err
	}
	return nil

}

// UpdateAll accepts a function that runs a transactional update on all rows
func (d *DataObj) UpdateAll(fn ExecFun) error {
	return fn(d.Conn)
}
