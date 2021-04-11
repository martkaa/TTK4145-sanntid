package distributor

import (
	"Project/cost"
	"Project/elevator"
	"Project/elevio"
)

type RequestState int // Bestillingens forskjellige tilstander,

const (
	None      RequestState = 0
	Order                  = 1
	Comfirmed              = 2
	Complete               = 3
)

type DistributorElevator struct { // Struct for heisen slik distributor modulen ser de
	Id       int
	Floor    int
	Dir      elevio.MotorDirection
	Requests [][]RequestState
	Behave   elevator.Behaviour
}

type DistributorOrder struct { // Struct man sende til og fra Cost.
	Elev DistributorElevator
	Req  elevio.ButtonEvent
}

func DistributorFsm(internalStateChan chan elevator.Behaviour, internalOrderChan chan elevio.ButtonEvent) {

	e := make([]*DistributorElevator, 0)
	elevators := e

	elevatorsUpdate := make(chan []DistributorElevator)
	newInternalRequest := make(chan elevio.ButtonEvent)
	assignedDistributorOrder := make(chan DistributorOrder) // Kanal for å motta bestilling fra Cost
	localChan := make(chan elevio.ButtonEvent)

	go elevio.PollButtons(newInternalRequest) // Innhenter nye lokale bestillinger fra elevio.

	for {
		select {

		case updatedElevators := <-elevatorsUpdate:
			distributorUpdate(elevators, updatedElevators)

		case r := <-newInternalRequest:
			go cost.Cost(elevators, r, assignedDistributorOrder)

		case D := <-assignedDistributorOrder:
			distributorOrderAssigned(D, localChan)

		case newBehaviour := <-internalStateChan:
			distributorUpdateInternalState(elevators, newBehaviour)
		}
	}
}

func distributorOrderAssigned(D DistributorOrder, localChan chan<- elevio.ButtonEvent) {
	if D.Elev.Id == localId {
		D.Elev.Requests[D.Req.Floor][D.Req.Button] = Comfirmed
		localChan <- D.Req //Fantastisk syntaks her!
	}
	/*else {
		Send to network
	}*/
}

func distributorUpdate(elevators []*DistributorElevator, updatedElevators []DistributorElevator) {
	for _, updatedElevator := range updatedElevators {
		for _, elevator := range elevators {
			if elevator.Id == updatedElevator.Id {
				*elevator = updatedElevator
				break
			}
		}
		elevators = append(elevators, &updatedElevator)
	}
}

func distributorUpdateInternalState(elevators []*DistributorElevator, updatedBehaviour elevator.Behaviour) {
	// Vi kan lage det sånn at den lokale heisen alltid har indeks 0?

	if len(elevators) == 0 {
		return
	}
	elevators[0].Behave = updatedBehaviour
}
