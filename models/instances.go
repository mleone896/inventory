package models

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq/hstore"
)

// InstanceConfigFun type allows function option configuration
type InstanceConfigFun func(*Instance)

// Instance encapsulates the ec2 table
type Instance struct {
	ID         int           `json:"id"`
	InstanceID string        `json:"instance_id"`
	AccountID  string        `json:"account_id"`
	SubnetID   string        `json:"subnet_id"`
	Tags       hstore.Hstore `json:"tags"`
}

// NewInstance constructor for a new color
func NewInstance(opts ...func(*Instance)) *Instance {
	col := &Instance{}

	// make sure we were passed something
	if len(opts) >= 1 {
		for _, opt := range opts {
			opt(col)
		}
	}

	return col
}

// WithDefaultInstance initializing blank server
func WithDefaultInstance() InstanceConfigFun {
	return func(c *Instance) {
	}
}

// WithColorTag ....
func WithColorTag(color string) InstanceConfigFun {
	return func(c *Instance) {
		c.Tags = hstore.Hstore{Map: addHstoreTag("color", color)}
	}
}

// GetAll satisfies finall
func (i *Instance) GetAll(db *sqlx.DB) (*sqlx.Rows, error) {
	sql := `SELECT * from ec2_instances`

	rows, err := db.Queryx(sql)

	if err != nil {
		return nil, fmt.Errorf("error retrieving instances: %s", err)
	}
	return rows, nil
}

// GetByID takes a color and returns an instance
func (i *Instance) GetByID(db *sqlx.DB) error {
	color, ok := i.Tags.Map["color"]

	if !ok {
		return fmt.Errorf("getByID: no color specified")
	}

	sql := `SELECT * from ec2_instances where ec2_instances.tags->'color' = $1`
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("could not get tx: %s", err)
	}

	if err := tx.QueryRowx(sql, color).StructScan(i); err != nil {
		return fmt.Errorf("could not find color in database: %s", err)
	}

	return nil
}

// Instances ...
func Instances() *Instance {
	inst := NewInstance(WithDefaultInstance())
	return inst
}

// Sync ...
func (i *Instance) Sync(db *sqlx.DB, instances []*Instance) error {

	sql := ` 
		INSERT INTO ec2_instances 
		(
		instance_id,
		account_id,
		subnet_id,
		tags
	    )
		VALUES (
			:instance_id, 
			:account_id, 
			:subnet_id,
			:tags)
			ON CONFLICT (instance_id, account_id) 
			DO UPDATE
			SET tags = :tags 
			WHERE ec2_instances.instance_id = :instance_id
			AND ec2_instances.account_id = :account_id
			`

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("could not get transaction: %s", err)
	}
	// prepare the parameterized statement
	stmt, err := tx.PrepareNamed(sql)

	if err != nil {
		return fmt.Errorf("could not prepare stmt: %s", err)
	}

	// loop over the instances and exec
	for _, instance := range instances {
		_, err := stmt.Exec(instance)

		if err != nil {
			txerr := tx.Rollback()
			if txerr != nil {
				return fmt.Errorf("error rollingback tx : %s", err)
			}
			return fmt.Errorf("stmt error rolling back tx: %s", err)
		}

	}
	err = tx.Commit()

	return nil
}

func addHstoreTag(key, tag string) map[string]sql.NullString {
	data := make(map[string]sql.NullString)
	data[key] = sql.NullString{String: tag, Valid: true}

	return data
}
