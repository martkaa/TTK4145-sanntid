package main

import (
	"Project/distributor/distributor"
	"Project/distributor/watchdog"
	"Project/elevator"
	"Project/elevio"
	"Project/fsm"
)

func main() {

	// Tenker at main blir den delen som "binder" sammen de forskjellige delene ved å lage forskjellige
	// kanaler og sende de inn i forskjellige go-rutiner.

	ch_internalOrderChan := make(chan elevio.ButtonEvent) //Channel for new internal orders
	ch_internalStateChan := make(chan elevator.Behaviour) //Channel for internal state

	go fsm.Fsm(ch_internalOrderChan, ch_internalStateChan)
	go distributor.DistributorFsm(ch_internalStateChan, ch_internalOrderChan)
	elevators := make([]*config.DistributorElevator, 0) 

	//Init watchdog
	watchdogTimeoutC := make(bool chan)
	watchdogUpdateStateC := make(chan elevators, 10)
	go watchdog.InitWatchdog(watchdogTimeoutC, watchdogUpdateStateC, config.WATCHDOG_TIMEOUT)
}
