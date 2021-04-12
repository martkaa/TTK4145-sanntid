package distributor

import (
	"Project/communication"
	"Project/cost"
	"Project/elevType"
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

const localId = 1

/* Input to cost module*/
type DistributorOrder struct {
	Elev elevType.Distributor
	Req  elevType.Request
}

func DistributorFsm(ch_internalStateChan chan elevator.Behaviour, ch_internalOrderChan chan elevio.ButtonEvent) {

	/*
		Communication stuff
	*/

	/* Channels for sending and receiving elevator struct*/
	ch_receive := make(chan peers.PeerUpdate)
	ch_transmit := make(chan elevType.Distributor)

	/* We can disable/enable the transmitter after it has been started.*/
	/* This could be used to signal that we are somehow "unavailable".*/
	ch_peerTxEnable := make(chan bool)
	ch_peerUpdate := make(chan peers.PeerUpdate)

	go communication.CommunicationInit(ch_receive, ch_transmit)
	go communication.PeerUpdateInit(ch_receive, ch_peerTxEnable)

	/**/

	/* Array containing all elevators on network*/
	elevators := make([]*elevType.Distributor, 0)

	/* Channel for triggers in fsm*/
	ch_elevatorsUpdate := make(chan []elevType.Distributor)
	ch_newInternalRequest := make(chan elevio.ButtonEvent)
	ch_assignedDistributorOrder := make(chan *elevType.Distributor) /* Channel for receiving assigned order from Cost */
	ch_localChan := make(chan elevio.ButtonEvent)

	go elevio.PollButtons(ch_newInternalRequest) /* Channel for receiving new local orders */

	for {
		select {

		case updatedElevators := <-ch_elevatorsUpdate:
			distributorUpdate(elevators, updatedElevators)

		case r := <-ch_newInternalRequest:
			go cost.Cost(elevators, r, ch_assignedDistributorOrder)

		case assignedElevator := <-ch_assignedDistributorOrder:
			// Konvertere fra elevator til distributorElevator
			updateDistributorElevators(elevators, *assignedElevator)

		case newBehaviour := <-ch_internalStateChan:
			distributorUpdateInternalState(elevators, newBehaviour)
		}
	}
}

func distributorOrderAssigned(order DistributorOrder, ch_localChan chan<- elevio.ButtonEvent) {
	if order.Elev.Id == localId {
		order.Elev.Requests[order.Req.Floor][order.Req.Button] = elevType.Comfirmed
		ch_localChan <- order.Req /* Take a look at this syntaks! */
	}
	/*else {
		Send to network
	}*/
}

func distributorUpdate(elevators []*elevType.Distributor, updatedElevators []elevType.Distributor) {
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

func distributorUpdateInternalState(elevators []*elevType.Distributor, updatedBehaviour elevator.Behaviour) {
	// Vi kan lage det sånn at den lokale heisen alltid har indeks 0?

	if len(elevators) == 0 {
		return
	}
	elevators[0].Behave = updatedBehaviour
}

/* Updating the DistributorElevators according to elevator assigned from Cost-function */

func updateDistributorElevators(elevators []*elevType.Distributor, assignedOrderElevator elevType.Distributor) {
	for _, e := range elevators {
		if e.Id == costElevator.Id {
			e.Requests[assignedOrderElevator.Floor][assignedOrderElevator.Button] = elevType.Comfirmed
		}
	}
}
