package controllers

import (
	"fmt"
	"sync"
	"time"

	shared "github.com/agile-work/srv-shared"
	"github.com/agile-work/srv-shared/amqp"
	"github.com/agile-work/srv-shared/sql-builder/builder"
	"github.com/agile-work/srv-shared/sql-builder/db"
)

//Scheduler defines an scheduler
type Scheduler struct {
	WG sync.WaitGroup
}

// JobInstance defines the struct of this object
type JobInstance struct {
	ID        string    `json:"id" sql:"id" pk:"true"`
	Status    string    `json:"status" sql:"status"`
	UpdatedAt time.Time `json:"updated_at" sql:"updated_at"`
}

// CheckJobsToExecute verify in the database instances to be executed and create a new message in the queue
func (s *Scheduler) CheckJobsToExecute(jobsQueue *amqp.Queue) {
	// TODO: Pensar em um jeito de impedir threads diferentes pegarem a mesma instância de job
	jobInstances := []JobInstance{}
	jobInstanceTable := shared.TableCoreJobInstances
	condition := builder.Equal("status", shared.JobStatusCreated)

	err := db.SelectStruct(jobInstanceTable, &jobInstances, condition)
	if err != nil {
		// TODO: Pensar em como tratar esse erro
		fmt.Println(err.Error())
	}

	for _, jobInstance := range jobInstances {
		jobInstanceIDColumn := fmt.Sprintf("%s.id", jobInstanceTable)
		condition := builder.Equal(jobInstanceIDColumn, jobInstance.ID)
		jobInstance.Status = shared.JobStatusInQueue
		jobInstance.UpdatedAt = time.Now()

		err := jobsQueue.Push(amqp.Message{
			ID:    jobInstance.ID,
			Queue: "jobs",
		})
		if err != nil {
			jobInstance.Status = shared.JobStatusCreated
			fmt.Printf("JOB Instance ID: %s | Error trying to send to queue: %s\n", jobInstance.ID, err.Error())
		} else {
			fmt.Printf("JOB Instance ID: %s | Sent to queue successfully\n", jobInstance.ID)
		}

		err = db.UpdateStruct(jobInstanceTable, &jobInstance, condition, "status", "updated_at")
		if err != nil {
			// TODO: Pensar em como tratar esse erro
			fmt.Println(err.Error())
		}
	}

	s.WG.Done()
}

// CheckServicesStatus verify and update services active status
func (s *Scheduler) CheckServicesStatus() {
	// statemant := builder.Update(
	// 	shared.TableCoreServices,
	// 	"active",
	// ).Values(
	// 	false,
	// ).Where(
	// 	builder.And(
	// 		builder.Equal("status", shared.JobStatusCreated),
	// 		builder.Equal("queue_at", nil),
	// 	),
	// )
	// err := db.Exec(statemant)
	// if err != nil {
	// 	// TODO: Pensar em como tratar esse erro
	// 	fmt.Println(err.Error())
	// }
	s.WG.Done()
}
