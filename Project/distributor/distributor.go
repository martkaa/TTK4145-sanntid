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
	Order                  = 1
	Comfirmed              = 2
	Complete               = 3
)

type DistributorElevator struct {
	Id       int
	Floor    int
	Dir      elevio.MotorDirection
	Requests [][]RequestState
	Behave   elevator.Behaviour
}

/* Input to cost module*/
type DistributorOrder struct {
	Elev DistributorElevator
	Req  elevio.ButtonEvent
}

func DistributorFsm(internalStateChan chan elevator.Behaviour, internalOrderChan chan elevio.ButtonEvent) {

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
			// Konvertere alle elevators til CostElevators.
			go cost.Cost(elevators, r, ch_assignedDistributorOrder)

		case D := <-ch_assignedDistributorOrder:
			// Konvertere fra elevator til distributorElevator
			distributorOrderAssigned(D, ch_localChan)

		case newBehaviour := <-ch_internalStateChan:
			distributorUpdateInternalState(elevators, newBehaviour)
		}
	}
}

func distributorOrderAssigned(D DistributorOrder, ch_localChan chan<- elevio.ButtonEvent) {
	if D.Elev.Id == localId {
		D.Elev.Requests[D.Req.Floor][D.Req.Button] = Comfirmed
		ch_localChan <- D.Req /* Take a look at this syntaks! */
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
