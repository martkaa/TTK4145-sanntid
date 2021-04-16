package main

import (
	"Project/distributor"
	"Project/elevio"
	"flag"
)

func main() {

	elevio.Init("localhost:50009", 4)

	/* Set id from command line using 'go run main.go -id=our_id'*/
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

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
