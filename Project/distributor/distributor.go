package distributor

import (
	"Project/config"
	"Project/elevator"
	"Project/elevio"
	"Project/fsm"
	"Project/network/bcast"
	"Project/network/peers"
	"fmt"
)

/* Set id from command line using 'go run main.go -id=our_id'*/
/*
var id string
flag.StringVar(&id, "id", "", "id of this peer")
flag.Parse()
*/
/* Input to cost module*/
type DistributorOrder struct {
	Elev config.DistributorElevator
	Req  config.Request
}

func DistributorFsm(id string) {

	/*
		Communication stuff
	*/

	/* We make a channel for receiving updates on the id's of the peers that are alive on the network*/
	ch_peerUpdate := make(chan peers.PeerUpdate)

	/* We can disable/enable the transmitter after it has been started. This could be used to signal that we are somehow "unavailable".*/
	ch_peerTxEnable := make(chan bool)

	go peers.Transmitter(15647, id, ch_peerTxEnable)
	go peers.Receiver(15647, ch_peerUpdate)

	/* Channels for sending and receiving elevator struct*/
	ch_receive := make(chan []config.DistributorElevator)
	ch_transmit := make(chan []config.DistributorElevator)

	go bcast.Transmitter(16569, ch_transmit)
	go bcast.Receiver(16569, ch_receive)

	/*
		elevators is an array containing all elevators on network
		Initialize elevators and set states
	*/

	/*p := ch_peerUpdate*/

	elevators := make([]*config.DistributorElevator, 0)
	thisElevator := distributorElevatorInit(id)
	elevators = append(elevators, &thisElevator)

	/* Update elevator in elevators that corresponds to local elevator*/

	/*
		If len(p.peers) > 1
		Broadcast elevator(s)
	*/

	/**/

	/* Channel for triggers in fsm*/
	//ch_elevatorsUpdate := make(chan []config.DistributorElevator)
	ch_newInternalRequest := make(chan elevio.ButtonEvent)
	ch_localElevUpdate := make(chan elevator.Elevator)
	//ch_assignedDistributorOrder := make(chan *config.DistributorElevator) /* Channel for receiving assigned order from Cost*/
	ch_orderToLocal := make(chan elevio.ButtonEvent) /* Channel for getting button event from fsm */

	go elevio.PollButtons(ch_newInternalRequest) /* Channel for receiving new local orders*/

	go fsm.Fsm(ch_orderToLocal, ch_localElevUpdate)

	for {
		select {

		/*
			case updatedElevators := <-ch_elevatorsUpdate:
				//Update local elevator in elevators
				// Broadcast
				distributorUpdate(elevators, updatedElevators)*/

		case r := <-ch_newInternalRequest:

			/* Check if hall or cab order*/
			/*
				If cab order, send to local elevator
			*/

			/* If hall order ...*/
			///go cost.Cost(elevators, r, ch_assignedDistributorOrder)
			//assignedElevator := <-ch_assignedDistributorOrder
			/* Check if order is assign to local or external elevator*/
			/*
				If local elevator, set corresponding element on elevator.Requests to confirmed and broadcast
			*/

			/* If external elevator, set corresponding element on elevator.Requests to Order ... */
			//updateDistributorElevators(elevators, *assignedElevator)
			/* Broadcast*/
			elevators[0].Requests[r.Floor][r.Button] = config.Comfirmed
			ch_orderToLocal <- r
			fmt.Println(elevators[0].Requests)
			distributorTransmit(elevators, ch_transmit)

		case updatedSate := <-ch_localElevUpdate:
			distributorUpdateInternalState(elevators, updatedSate)
			distributorTransmit(elevators, ch_transmit)

		case updatedElevators := <-ch_receive:
			distributorUpdateElevators(elevators, updatedElevators)
		}
	}
}

/*
	Elevator-state update stuff
*/
func distributorUpdateElevators(elevators []*config.DistributorElevator, updatedElevators []config.DistributorElevator) {
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

func distributorUpdateInternalState(elevators []*config.DistributorElevator, e elevator.Elevator) {
	// Vi kan lage det sånn at den lokale heisen alltid har indeks 0?
	elevators[0].Behave = config.Behaviour(e.Behave)
	elevators[0].Floor = e.Floor
	for floor := range e.Requests {
		for button := range e.Requests[floor] {
			if !e.Requests[floor][button] && elevators[0].Requests[floor][button] == config.Comfirmed {
				elevators[0].Requests[floor][button] = config.Complete
			}
		}
	}
}

/*
	Assigning order stuff
*/

/*func distributorOrderAssigned(order DistributorOrder, ch_localChan chan<- elevio.ButtonEvent) {
if order.Elev.Id ==  {
	order.Elev.Requests[order.Req.Floor][order.Req.Button] = config.Comfirmed
	ch_localChan <- order.Req /* Take a look at this syntaks! */
/*else {
		Send to network
	}
}*/

/* Updating the DistributorElevators according to elevator assigned from Cost-function */

func distributorTransmit(elevators []*config.DistributorElevator, ch_transmit chan<- []config.DistributorElevator) {
	tempElevators := make([]config.DistributorElevator, 0)
	for _, e := range elevators {
		tempElevators = append(tempElevators, *e)
	}
	ch_transmit <- tempElevators
}

func distributorElevatorInit(id string) config.DistributorElevator {
	requests := make([][]config.RequestState, 4)
	for floor := range requests {
		requests[floor] = make([]config.RequestState, 3)
	}
	return config.DistributorElevator{Requests: requests, Id: id, Floor: 0, Behave: config.Idle}
}
