package controllers

import (
	"fmt"
	"sync"
	"time"

	"github.com/agile-work/srv-shared/rdb"

	"github.com/agile-work/srv-shared/constants"
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

// CheckJobsToExecute verify in the database instances to be executed and create a new message in the redis queue
func (s *Scheduler) CheckJobsToExecute(now time.Time) {
	if rdb.Available() {
		// TODO: Pensar em um jeito de impedir threads diferentes pegarem a mesma instância de job
		jobInstances := []JobInstance{}
		jobInstanceTable := constants.TableCoreJobInstances

		err := db.SelectStruct(jobInstanceTable, &jobInstances, &db.Options{
			Conditions: builder.Equal("status", constants.JobStatusCreated),
		})
		if err != nil {
			// TODO: Pensar em como tratar esse erro
			fmt.Println(err.Error())
			s.WG.Done()
			return
		}

		jobInstanceIDColumn := fmt.Sprintf("%s.id", jobInstanceTable)

		for _, jobInstance := range jobInstances {
			jobInstance.Status = constants.JobStatusInQueue
			jobInstance.UpdatedAt = time.Now()

			err := rdb.LPush("queue:jobs", jobInstance.ID)

			if err != nil {
				jobInstance.Status = constants.JobStatusCreated
				fmt.Printf("JOB Instance ID: %s | Error trying to send to queue: %s\n", jobInstance.ID, err.Error())
				continue
			} else {
				fmt.Printf("JOB Instance ID: %s | Sent to queue successfully\n", jobInstance.ID)
			}

			// TODO: change to transaction
			err = db.UpdateStruct(jobInstanceTable, &jobInstance, &db.Options{
				Conditions: builder.Equal(jobInstanceIDColumn, jobInstance.ID),
			}, "status", "updated_at")
			if err != nil {
				// TODO: Pensar em como tratar esse erro
				fmt.Println(err.Error())
			}
		}
	}
	s.WG.Done()
}
