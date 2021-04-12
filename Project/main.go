package main

import (
	"Project/distributor"
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

}
