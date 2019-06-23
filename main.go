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
	"github.com/agile-work/srv-shared/service"
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

	pid := os.Getpid()
	hostname, _ := os.Hostname()
	aux := service.New("Scheduler", constants.ServiceTypeAuxiliary, hostname, 0, pid)

	fmt.Printf("Starting Service %s...\n", aux.Name)
	fmt.Printf("[Instance: %s | PID: %d]\n", aux.InstanceCode, aux.PID)

	fmt.Println("Database connecting...")
	err := db.Connect(host, port, user, password, dbName, false)
	if err != nil {
		fmt.Println("Error connecting to database")
		return
	}
	fmt.Println("Database connected")

	rdb.Init(*redisHost, *redisPort, *redisPass)
	defer rdb.Close()

	socket.Init(aux, *wsHost, *wsPort)
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

	fmt.Println("Scheduler ready...")

	<-stopChan
	fmt.Println("\nShutting down Service...")
	fmt.Println("Service stopped!")
}
