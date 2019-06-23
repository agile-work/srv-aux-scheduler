package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/agile-work/srv-aux-scheduler/controllers"
	"github.com/agile-work/srv-shared/constants"
	"github.com/agile-work/srv-shared/rdb"
	"github.com/agile-work/srv-shared/socket"
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
	redisHost           = flag.String("redisHost", "localhost", "Redis host")
	redisPort           = flag.Int("redisPort", 6379, "Redis port")
	redisPass           = flag.String("redisPass", "redis123", "Redis password")
	wsHost              = flag.String("wsHost", "localhost", "Realtime host")
	wsPort              = flag.Int("wsPort", 8010, "Realtime port")
)

func main() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	flag.Parse()
	fmt.Printf("Starting Service %s...\n", *serviceInstanceName)

	fmt.Println("Database connecting...")
	err := db.Connect(host, port, user, password, dbName, false)
	if err != nil {
		fmt.Println("Error connecting to database")
		return
	}
	fmt.Println("Database connected")

	rdb.Init(*redisHost, *redisPort, *redisPass)
	defer rdb.Close()

	socket.Init(*serviceInstanceName, constants.ServiceTypeAuxiliary, *wsHost, *wsPort)
	defer socket.Close()

	scheduler := controllers.Scheduler{}

	ticker := time.NewTicker(time.Duration(*execInterval) * time.Second)
	go func() {
		for t := range ticker.C {
			scheduler.WG.Add(1)
			scheduler.CheckJobsToExecute(t)
			scheduler.WG.Wait()
		}
	}()

	fmt.Printf("Scheduler pid:%d ready...\n", os.Getpid())

	<-stopChan
	fmt.Println("\nShutting down Service...")
	fmt.Println("Service stopped!")
}
