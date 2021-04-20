package main

import (
	"flag"

	"Project/distributor"

	"Project/elevio"
)

func main() {

	/* Set port from command line using 'go run main.go -port=our_id'*/
	var port string
	flag.StringVar(&port, "port", "", "port of this peer")

	/* Set id from command line using 'go run main.go -id=our_id'*/
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	elevio.Init("localhost:"+port, 4)

	finished := make(chan int)

	// Tenker at main blir den delen som "binder" sammen de forskjellige delene ved å lage forskjellige
	// kanaler og sende de inn i forskjellige go-rutiner.

	go distributor.DistributorFsm(id)

	<-finished
	//Init watchdog
	/*
		watchdogTimeoutC := make(chan bool)
		watchdogUpdateStateC := make(chan elevators, 10)
		go watchdog.InitWatchdog(watchdogTimeoutC, watchdogUpdateStateC, config.WATCHDOG_TIMEOUT)
	*/
}
