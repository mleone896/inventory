package runners

import (
	"fmt"
	"log"
	"time"
)

// Runner ...
type Runner interface {
	Exec(JobExecFunc) error
}

// RunConfigFunc ...
type RunConfigFunc func(*Run) error

// Run ...
type Run struct {
	pollInterval int
	done         chan struct{}
	Job          *Job
	Desc         string
}

// New ...
func New(options ...RunConfigFunc) (*Run, error) {
	run := &Run{
		pollInterval: 30, // set a default
		done:         make(chan struct{}),
	}

	for _, option := range options {
		if err := option(run); err != nil {
			return nil, fmt.Errorf("could not set run function option %s", err)
		}
	}

	return run, nil

}

// WithDescription ...
func WithDescription(s string) RunConfigFunc {
	return func(r *Run) error {
		r.Desc = s
		return nil
	}
}

// Loop creates a standard template for calling a function with a given interval
func (r *Run) Loop(rf func(*Job) error) {
	ticker := time.NewTicker(time.Duration(r.pollInterval) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Printf("executing %s loop", r.Desc)
				if err := rf(r.Job); err != nil {
					log.Println("fatal error stopping job %v", err)
					r.Stop()
				}
			case <-r.done:
				ticker.Stop()
				return
			}

		}
	}()

}

// WithInterval takes the poll rate in seconds and returns a jobConfigFunc type that can
// be passed into New
func WithInterval(i int) RunConfigFunc {
	return func(j *Run) error {
		j.pollInterval = i
		return nil
	}
}

// WithJob  sets the function that will be run as a task
func WithJob(rf *Job) RunConfigFunc {
	return func(j *Run) error {
		j.Job = rf
		return nil
	}
}

// Stop will send a message to shutdown the task running
func (r *Run) Stop() {
	close(r.done)
}
