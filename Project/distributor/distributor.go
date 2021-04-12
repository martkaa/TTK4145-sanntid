package distributor

import (
	"Project/communication"
	"Project/cost"
	"Project/elevator"
	"Project/elevio"
	"Project/network/peers"
)

/* Set id from command line using 'go run main.go -id=our_id'*/
/*
var id string
flag.StringVar(&id, "id", "", "id of this peer")
flag.Parse()
*/

type RequestState int

/* Order types*/
const (
	None      RequestState = 0
	Order     RequestState = 1
	Comfirmed RequestState = 2
	Complete  RequestState = 3
)

type DistributorElevator struct {
	Id       int
	Floor    int
	Dir      elevio.MotorDirection
	Requests [elevator.NumFloors][elevator.NumButtons]RequestState
	Behave   elevator.Behaviour
}

/* Input to cost module*/
type DistributorOrder struct {
	Elev DistributorElevator
	Req  elevio.ButtonEvent
}

func DistributorFsm(ch_internalStateChan chan elevator.Behaviour, ch_internalOrderChan chan elevio.ButtonEvent) {

	/*
		Communication stuff
	*/

	/* Channels for sending and receiving elevator struct*/
	ch_receive := make(chan peers.PeerUpdate)
	ch_transmit := make(chan DistributorElevator)

	/* We can disable/enable the transmitter after it has been started.*/
	/* This could be used to signal that we are somehow "unavailable".*/
	ch_peerTxEnable := make(chan bool)
	ch_peerUpdate := make(chan peers.PeerUpdate)

	go communication.CommunicationInit(ch_receive, ch_transmit)
	go communication.PeerUpdateInit(ch_receive, ch_peerTxEnable)

	/**/

	/* Array containing all elevators on network*/
	e := make([]*DistributorElevator, 0)
	elevators := e

	/* Channel for triggers in fsm*/
	ch_elevatorsUpdate := make(chan []DistributorElevator)
	ch_newInternalRequest := make(chan elevio.ButtonEvent)
	ch_assignedDistributorOrder := make(chan cost.CostElevator) /* Channel for receiving assigned order from Cost */
	ch_localChan := make(chan elevio.ButtonEvent)

	go elevio.PollButtons(ch_newInternalRequest) /* Channel for receiving new local orders */

	for {
		select {

		case updatedElevators := <-ch_elevatorsUpdate:
			distributorUpdate(elevators, updatedElevators)

		case r := <-ch_newInternalRequest:
			costElevators := distributorElevatorsToCostElevators(elevators, r) /* Converting from DistributorElevatosr to CostElevators */
			go cost.Cost(costElevators, r, ch_assignedDistributorOrder)

		case costElevator := <-ch_assignedDistributorOrder:
			// Konvertere fra elevator til distributorElevator
			updateDistributorElevators(elevators, costElevator)

		case newBehaviour := <-ch_internalStateChan:
			distributorUpdateInternalState(elevators, newBehaviour)
		}
	}
}

func distributorOrderAssigned(order DistributorOrder, ch_localChan chan<- elevio.ButtonEvent) {
	if order.Elev.Id == localId {
		order.Elev.Requests[order.Req.Floor][order.Req.Button] = Comfirmed
		ch_localChan <- order.Req /* Take a look at this syntaks! */
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

/* Updating the DistributorElevators according to elevator assigned from Cost-function */

func updateDistributorElevators(elevators []*DistributorElevator, costElevator cost.CostElevator) {
	for _, e := range elevators {
		if e.Id == costElevator.Id {
			e.Requests[costElevator.Req.Floor][costElevator.Req.Button] = Comfirmed
		}
	}
}

/* Convert from distributorElevator to CostElevators befor sending to Cost-module*/

func distributorElevatorsToCostElevators(elevators []*DistributorElevator, r elevio.ButtonEvent) []cost.CostElevator {
	var costElevators []cost.CostElevator
	for _, e := range elevators {
		elevator := elevator.Elevator{
			Floor:    e.Floor,
			Dir:      e.Dir,
			Requests: distributorRequestsToElevatorRequest(e.Requests),
			Behave:   e.Behave,
		}
		costElevators = append(costElevators, cost.CostElevator{
			Id:   e.Id,
			Elev: elevator,
			Req:  r})
	}
	return costElevators
}

func distributorRequestsToElevatorRequest(distributorRequests [elevator.NumFloors][elevator.NumButtons]RequestState) [elevator.NumFloors][elevator.NumButtons]bool {
	var elevatorRequests [elevator.NumFloors][elevator.NumButtons]bool
	for floor := range distributorRequests {
		for button := range distributorRequests[floor] {
			if distributorRequests[floor][button] == Comfirmed {
				elevatorRequests[floor][button] = true
			}
		}
	}
	return elevatorRequests
}

/* Function converting from CommunicationElevator to DistributorElevator */
func communicationElevatorToDistributorElevator(c communication.CommunicationElevator) DistributorElevator {
	e := DistributorElevator{
		Id:     c.Id,
		Floor:  c.Floor,
		Dir:    c.Dir,
		Behave: c.Behave}
	for floor := range e.Requests {
		for button := range e.Requests[floor] {
			e.Requests[floor][button] = RequestState(c.Requests[floor][button])
		}
	}
	return e
}
