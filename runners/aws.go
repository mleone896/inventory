package runners

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/lib/pq/hstore"
	"github.com/mleone896/inventory/models"
)

// Conn  .....
type Conn struct {
	ec2       *ec2.EC2
	accountID string
}

// NewConn ...
func NewConn(region, aid string) (*Conn, error) {
	c := new(Conn)

	creds := credentials.NewEnvCredentials()

	c.ec2 = ec2.New(session.New(), &aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	})

	c.accountID = aid

	return c, nil
}

func (c *Conn) getInstances() ([]*models.Instance, error) {

	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("instance-state-name"),
				Values: []*string{aws.String("running")},
			},
		},
	}

	resp, err := c.ec2.DescribeInstances(params)

	if err != nil {
		return nil, fmt.Errorf("could not describe instances: %s", err)
	}

	instances := []*models.Instance{}
	for idx := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {

			data := make(map[string]sql.NullString)
			for _, val := range inst.Tags {
				data[*val.Key] = sql.NullString{String: *val.Value, Valid: true}
			}
			record := &models.Instance{
				InstanceID: *inst.InstanceId,
				AccountID:  c.accountID,
				SubnetID:   *inst.SubnetId,
				Tags:       hstore.Hstore{Map: data},
			}

			instances = append(instances, record)

		}
	}
	log.Println("about to exit getInstances")

	return instances, nil
}

func (c *Conn) colorsFromTags(instances []models.Instance) []string {
	colors := make([]string, len(instances))

	for idx, instance := range instances {
		if _, ok := instance.Tags.Map["color"]; ok {
			colors[idx] = instance.Tags.Map["color"].String
		}

	}

	return colors
}

func (c *Conn) getsubnets() ([]*models.Subnet, error) {
	subs := []*models.Subnet{}
	resp, err := c.ec2.DescribeSubnets(nil)

	if err != nil {
		return subs, fmt.Errorf("could not describe subnets: %s", err)
	}

	for _, sub := range resp.Subnets {
		subnet := &models.Subnet{
			SubnetID:  *sub.SubnetId,
			VpcID:     *sub.VpcId,
			AZ:        *sub.AvailabilityZone,
			Tags:      hstore.Hstore{Map: convertTags(sub.Tags)},
			AccountID: c.accountID,
		}

		subs = append(subs, subnet)

	}

	return subs, nil

}

func convertTags(tags []*ec2.Tag) map[string]sql.NullString {
	data := make(map[string]sql.NullString)

	for _, tag := range tags {
		data[*tag.Key] = sql.NullString{String: *tag.Value, Valid: true}

	}
	return data
}
