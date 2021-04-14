package distributor

import (
	"Project/config"
	"Project/cost"
	"Project/elevator"
	"Project/elevio"
	"Project/network/peers"
	"flag"
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
	Elev config.DistributorElevator
	Req  config.Request
}

func DistributorFsm(ch_internalStateChan chan elevator.Behaviour, ch_internalOrderChan chan elevio.ButtonEvent) {

	/*
		Communication stuff
	*/

	/* Set id from command line using 'go run main.go -id=our_id'*/

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	/* We make a channel for receiving updates on the id's of the peers that are alive on the network*/
	ch_peerUpdate := make(chan peers.PeerUpdate)

	/* We can disable/enable the transmitter after it has been started. This could be used to signal that we are somehow "unavailable".*/
	ch_peerTxEnable := make(chan bool)

	go peers.Transmitter(15647, id, ch_peerTxEnable)
	go peers.Receiver(15647, ch_peerUpdate)

	/* Channels for sending and receiving elevator struct*/
	ch_receive := make(chan config.DistributorElevator)
	ch_transmit := make(chan config.DistributorElevator)

	go bcast.Transmitter(16569, ch_transmit)
	go bcast.Receiver(16569, ch_receive)

	/*
		elevators is an array containing all elevators on network
		Initialize elevators and set states
	*/

	p := ch_peerUpdate

	elevators := make([]*config.DistributorElevator, len(p.Peers))

	/* Update elevator in elevators that corresponds to local elevator*/

	/*
		If len(p.peers) > 1
		Broadcast elevator(s)
	*/

	/**/

	/* Channel for triggers in fsm*/
	ch_elevatorsUpdate := make(chan []config.DistributorElevator)
	ch_newInternalRequest := make(chan elevio.ButtonEvent)
	ch_assignedDistributorOrder := make(chan *config.DistributorElevator) /* Channel for receiving assigned order from Cost*/
	ch_localChan := make(chan elevio.ButtonEvent)

	go elevio.PollButtons(ch_newInternalRequest) /* Channel for receiving new local orders*/

	for {
		select {

		case updatedElevators := <-ch_elevatorsUpdate:
			/* Update local elevator in elevators*/
			/* Broadcast*/
			distributorUpdate(elevators, updatedElevators)

		case r := <-ch_newInternalRequest:
			/* Check if hall or cab order*/
			/*
				If cab order, send to local elevator
			*/

			/* If hall order ...*/
			go cost.Cost(elevators, r, ch_assignedDistributorOrder)

		case assignedElevator := <-ch_assignedDistributorOrder:
			/* Check if order is assign to local or external elevator*/
			/*
				If local elevator, set corresponding element on elevator.Requests to confirmed and broadcast
			*/

			/* If external elevator, set corresponding element on elevator.Requests to Order ... */
			updateDistributorElevators(elevators, *assignedElevator)
			/* Broadcast*/

		case newBehaviour := <-ch_internalStateChan:
			distributorUpdateInternalState(elevators, newBehaviour)

		case e := ch_receive:
			for _, elevator := range elevators {
				if e.Id == elevator.Id {
					elevator = e
					break
				}
			}
		}
	}
}

/*
	Elevator-state update stuff
*/
func distributorUpdate(elevators []*config.DistributorElevator, updatedElevators []config.DistributorElevator) {
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

func distributorUpdateInternalState(elevators []*config.DistributorElevator, updatedBehaviour elevator.Behaviour) {
	// Vi kan lage det sånn at den lokale heisen alltid har indeks 0?

	if len(elevators) == 0 {
		return
	}
	elevators[0].Behave = updatedBehaviour
}

/*
	Assigning order stuff
*/

func distributorOrderAssigned(order DistributorOrder, ch_localChan chan<- elevio.ButtonEvent) {
	if order.Elev.Id == localId {
		order.Elev.Requests[order.Req.Floor][order.Req.Button] = config.Comfirmed
		ch_localChan <- order.Req /* Take a look at this syntaks! */
	}
	/*else {
		Send to network
	}*/
}

/* Updating the DistributorElevators according to elevator assigned from Cost-function */

func updateDistributorElevators(elevators []*config.DistributorElevator, assignedOrderElevator config.DistributorElevator) {
	for _, e := range elevators {
		if e.Id == costElevator.Id {
			e.Requests[assignedOrderElevator.Floor][assignedOrderElevator.Button] = config.Order
		}
	}
}
