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
	"github.com/agile-work/srv-shared/sql-builder/db"
)

var (
	serviceInstanceName = flag.String("name", "Scheduler", "Name of this instance")
	heartbeatInterval   = flag.Int("heartbeat", 10, "Number of seconds to send a heartbeat")
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

	service, err := shared.RegisterService(*serviceInstanceName, shared.ServiceTypeAuxiliary)
	if err != nil {
		fmt.Println("Error registering service in the database")
		return
	}
	fmt.Printf("Service %s registered\n", service.ID)

	jobsQueue, _ := amqp.New("amqp://guest:guest@localhost:5672/", "jobs", false)

	scheduler := controllers.Scheduler{}

	ticker := time.NewTicker(time.Duration(*heartbeatInterval) * time.Second)
	go func() {
		for t := range ticker.C {
			service.Heartbeat(t)
			scheduler.WG.Add(2)
			go scheduler.CheckJobsToExecute(jobsQueue)
			go scheduler.CheckServicesStatus()
			scheduler.WG.Wait()
		}
	}()

	<-stopChan
	fmt.Println("Shutting down Service...")
	service.Down()
	fmt.Println("Service stopped!")
}
