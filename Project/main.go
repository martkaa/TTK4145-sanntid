package main

import (
	"Project/distributor"
	"Project/elevator"
	"Project/fsm"
)

func main() {

	internalOrderChan := make(chan distributor.Request) //Channel for new internal orders
	internalStateChan := make(chan elevator.Elevator)   //Channel for internal state

	go fsm.Fsm(internalOrderChan, internalStateChan)
	go distributor.DistributorFsm(internalStateChan, internalOrderChan)

}
