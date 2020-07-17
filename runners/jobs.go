package runners

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/mleone896/inventory/models"
)

// Job ...
type Job struct {
	aws *Conn
	db  *sqlx.DB
	aid string
}

// JobConfigFunc ...
type JobConfigFunc func(*Job) error

// JobExecFunc ...
type JobExecFunc func(*Job) error

// NewJob ...
func NewJob(options ...JobConfigFunc) (*Job, error) {
	job := &Job{}

	for _, option := range options {
		if err := option(job); err != nil {
			return nil, fmt.Errorf("could not set function option %s", err)
		}
	}

	return job, nil
}

// WithDataBase ...
func WithDataBase(db *sqlx.DB) JobConfigFunc {
	return func(j *Job) error {
		j.db = db
		return nil
	}

}

// WithAccount ...
func WithAccount(aid string) JobConfigFunc {
	return func(j *Job) error {
		j.aid = aid
		return nil
	}

}

// WithAwsConnection connects to aws and returns a conn object
func WithAwsConnection(region, aid string) JobConfigFunc {

	return func(j *Job) error {
		c, err := NewConn(region, aid)
		j.aid = aid
		if err != nil {
			return fmt.Errorf("could not get aws connection %s", err)
		}
		j.aws = c
		return nil
	}
}

// Exec takes a function that adheres to the JobExecFunc type
func (j *Job) Exec(jf JobExecFunc) error {
	err := jf(j)
	if err != nil {
		return err
	}
	return nil
}

// PopulateSubnets ...
func PopulateSubnets(j *Job) error {

	log.Println("populateSubnets: retrieving subnets from aws")
	subnets, err := j.aws.getsubnets()

	if err != nil {
		return fmt.Errorf("could not get subnets from AWS: %s", err)
	}

	account := models.NewAccount(j.aid)
	log.Println("populateSubnets: syncing aws subnets")
	if err := account.Sync(j.db, subnets); err != nil {
		log.Println(err)
		return err
	}

	log.Println("populateSubnets: complete")
	return nil
}

// PopulateInstances ...
func PopulateInstances(j *Job) error {

	log.Println("populateInstances: retrieving instances from aws")
	instances, err := j.aws.getInstances()
	if err != nil {
		return err
	}

	log.Println("populateInstances: syncing aws instances")
	inst := models.Instances()

	err = inst.Sync(j.db, instances)
	if err != nil {
		log.Println(err)
		return err
	}

	colors := models.Colors()

	err = colors.Sync(j.db, colorsFromTags(instances))

	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("populateInstances: completed")

	return nil
}

// colorsFromTags returns a list of colors from ec2 information
func colorsFromTags(instances []*models.Instance) []string {
	colors := make([]string, len(instances))

	for idx, instance := range instances {
		if _, ok := instance.Tags.Map["color"]; ok {
			colors[idx] = instance.Tags.Map["color"].String
		}

	}

	return colors
}
