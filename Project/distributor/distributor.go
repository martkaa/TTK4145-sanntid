package distributor

import (
	"Project/elevator"
	"Project/elevio"
)

type RequestState int

const (
	None      RequestState = 0
	Order                  = 1
	Comfirmed              = 2
	Complete               = 3
)

type Request struct {
	Floor        int
	Btn          elevio.ButtonType
	DesignatedId int
}

type DistributorOrder struct {
	E elevator.Elevator
	R Request
}

func DistributorFsm(internalStateChan chan elevator.Elevator, internalOrderChan chan Request) {

	e := make([]elevator.Elevator, 0)
	elevators := &e

	elevatorsUpdate := make(chan []elevator.Elevator)
	newInternalRequest := make(chan Request)
	assignedDistributorOrder := make(chan DistributorOrder)

	for {
		select {
		case updatedElevators := <-elevatorsUpdate:
			distributorUpdate(elevators, updatedElevators)
		case r := <-newInternalRequest:
			go cost.Cost(*elevators, r, assignedDistributorOrder)
		case D := <-assignedDistributorOrder:
			distributorOrderAssigned(e, localChan)
		}
	}
}

func distributorOrderAssigned(e elevator.Elevator, localChan chan<- Request) {
	if e.Id == localId {
		localChan <- Request{}
	}
	/*else {
		Send to network
	}*/
}

func distributorUpdate(elevators *[]elevator.Elevator, updatedElevators []elevator.Elevator) {
	for _, updatedElevator := range updatedElevators {
		for _, elevator := range *elevators {
			if elevator.Id == updatedElevator.Id {
				elevator = updatedElevator
				break
			}
		}
		*elevators = append(*elevators, updatedElevator)
	}
}
