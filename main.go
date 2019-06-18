package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/agile-work/srv-aux-scheduler/controllers"

	shared "github.com/agile-work/srv-shared"
	"github.com/agile-work/srv-shared/amqp"
	"github.com/agile-work/srv-shared/service"
	"github.com/agile-work/srv-shared/sql-builder/db"
)

var (
	serviceInstanceName = flag.String("name", "Scheduler", "Name of this instance")
	execInterval        = flag.Int("execInterval", 10, "Interval (seconds) between scheduler executions")
	host                = "cryo.cdnm8viilrat.us-east-2.rds-preview.amazonaws.com"
	port                = 5432
	user                = "cryoadmin"
	password            = "x3FhcrWDxnxCq9p"
	dbName              = "cryo"
)

func main() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	flag.Parse()
	fmt.Println("Starting Service...")
	err := db.Connect(host, port, user, password, dbName, false)
	if err != nil {
		fmt.Println("Error connecting to database")
		return
	}
	fmt.Println("Database connected")

	jobsQueue, err := amqp.New("amqp://guest:guest@localhost:5672/", "jobs", false)
	if err != nil {
		fmt.Println("Error connecting to queue")
		return
	}
	fmt.Println("Queue connected")

	srv, err := service.Register(*serviceInstanceName, shared.ServiceTypeAuxiliary)
	if err != nil {
		fmt.Println("Error registering service in the database")
		return
	}
	fmt.Printf("Service %s registered\n", srv.ID)

	scheduler := controllers.Scheduler{}

	ticker := time.NewTicker(time.Duration(*execInterval) * time.Second)
	go func() {
		for t := range ticker.C {
			scheduler.CheckJobsToExecute(jobsQueue)
		}
	}()

	<-stopChan
	fmt.Println("Shutting down Service...")
	srv.Down()
	fmt.Println("Service stopped!")
}
