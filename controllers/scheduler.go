package controllers

import (
	"sync"

	"github.com/agile-work/srv-shared/amqp"
)

//Scheduler defines an scheduler
type Scheduler struct {
	WG sync.WaitGroup
}

//CheckJobsToExecute verify in the database instances to be executed and create a new message in the queue
func (s *Scheduler) CheckJobsToExecute(jobsQueue *amqp.Queue) {
	//TODO: Load available to execute JOB instances
	s.WG.Done()
}

//CheckServicesStatus verify and update services active status
func (s *Scheduler) CheckServicesStatus() {
	//TODO: Create a message to the job queue
	s.WG.Done()
}
