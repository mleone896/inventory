package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/mleone896/inventory/db"
	"github.com/mleone896/inventory/runners"
	"github.com/mleone896/inventory/server"
)

// DefaultPollInterval ...
const (
	DefaultPollInterval = 60
	DefaultPort         = ":8080"
	DefaultConnString   = "dbname=inventory sslmode=disable"
	DefaultAccount      = "181657471068"
	DefaultRegion       = "us-east-1"
)

var (
	pollInterval int
	connString   string
	port         string
	account      string
	region       string
)

func init() {
	flag.StringVar(&connString, "connString", DefaultConnString, "The Database ConnectionString")
	flag.StringVar(&port, "port", DefaultPort, "The default port to listen")
	flag.IntVar(&pollInterval, "pollInterval", DefaultPollInterval, "Poll Interval in seconds")
	flag.StringVar(&account, "account", DefaultAccount, "The aws account you're polling")
	flag.StringVar(&region, "region", DefaultRegion, "AWS region")
}

func main() {

	flag.Parse()

	d, err := db.New(db.WithConnString(connString))

	checkError(err, "db.New()")

	job, err := runners.NewJob(
		runners.WithAwsConnection(region, account),
		runners.WithDataBase(d.Conn),
	)

	checkError(err, "runners.NewJob(AWS Population Job)")

	runInstances, err := runners.New(
		runners.WithInterval(pollInterval),
		runners.WithDescription("AWS Population Job Instances"),
		runners.WithJob(job),
	)
	checkError(err, "runners.New()")

	runSubnets, err := runners.New(
		runners.WithInterval(pollInterval),
		runners.WithDescription("AWS Population Job Subnets"),
		runners.WithJob(job),
	)

	checkError(err, "runners.New()")

	log.Println("Initiating instances routine")

	runSubnets.Loop(runners.PopulateSubnets)
	runInstances.Loop(runners.PopulateInstances)

	// before starting the api wait  for things to be retrieved
	time.Sleep(30 * time.Second)

	//  create api server
	server := server.New(server.WithDAO(d))

	router := server.LoadHandlers()

	log.Println("Serving new inventory connections")
	log.Fatal(http.ListenAndServe(port, router))

}

// helper function
func checkError(err error, function string) {
	if err != nil {
		log.Printf("recived error from fn %s: err: %s", function, err)
	}

}
