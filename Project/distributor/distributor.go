package distributor

import (
	"../config"
	"../elevator"
	"../elevio"
	"../network/bcast"
)

const localElevator = 0

/* Set id from commasnd line using 'go run main.go -id=our_id'*/

func elevatorInit(id string) config.DistributorElevator {
	requests := make([][]config.RequestState, 4)
	for floor := range requests {
		requests[floor] = make([]config.RequestState, 3)
	}
	return config.DistributorElevator{Requests: requests, ID: id, Floor: 0, Behave: config.Idle}
}

func broadcast(elevators []*config.DistributorElevator, ch_transmit chan<- []config.DistributorElevator) {
	tempElevators := make([]config.DistributorElevator, 0)
	for _, elev := range elevators {
		tempElevators = append(tempElevators, *elev)
	}
	ch_transmit <- tempElevators
}

func distributorFsm(id string) {

	/* Channels */
	ch_newLocalOrder := make(chan elevio.ButtonEvent)
	ch_newLocalState := make(chan elevator.Elevator)
	ch_msgFromNetwork := make(chan []config.DistributorElevator)
	ch_msgToNetwork := make(chan []config.DistributorElevator)
	ch_orderToLocal := make(chan elevio.ButtonEvent)

	/* Functions for network communication */
	go bcast.Transmitter(16569, ch_msgToNetwork)
	go bcast.Receiver(16569, ch_msgFromNetwork)

	/* Initialiaze elevator */
	elevators := make([]*config.DistributorElevator, 0)
	thisElevator := elevatorInit(id)
	elevators = append(elevators, &thisElevator)

	for {
		select {
		case newOrder := <-ch_newLocalOrder:
			assignOrder(elevators, newOrder)
			if elevators[localElevator].Requests[newOrder.Floor][newOrder.Button] == config.Comfirmed {
				setHallLights(elevators)
				ch_orderToLocal <- newOrder
			}
			broadcast(elevators, ch_msgToNetwork)

		case newState := <-ch_newLocalState:
			updateLocalState(elevators, newState)
			checkOrderComplete(elevators, newState)
			broadcast(elevators, ch_msgToNetwork)

		case newElevators := <-ch_msgFromNetwork:
			updateElevators(elevators, newElevators)
			setHallLights(elevators)
			localComfirmedOrders := getComfirmedOrders(elevators[0].Requests)

			/* Må finne en måte å huske hvilken bestilling som er ny for å kunne sende akkurat denne bestillingen videre til lokal */

			if len(localComfirmedOrders) > 0 {
				ch_newLocalOrder <- elevio.ButtonEvent
				broadcast(elevators, ch_msgToNetwork)
			}
		}
	}
}

func getComfirmedOrders(localRequests [][]config.RequestState) []config.Request {
	var orders []config.Request
	for floor := range localRequests {
		for button := range localRequests[floor] {
			if localRequests[floor][button] == config.Order || localRequests[floor][button] == config.Comfirmed {
				order := config.Request{
					Floor:  floor,
					Button: config.ButtonType(button),
				}
				orders = append(orders, order)
			}
		}
	}
	return orders
}

/*
	New order from local stuff
*/

func assignOrder(elevators []*config.DistributorElevator, order elevio.ButtonEvent) {

	minCost := 99999
	cost := 0
	var minElev *config.DistributorElevator

	for _, elev := range elevators {
		cost = cost.Cost(*elev)
		if cost < minCost {
			minCost = cost
			minElev = elev
		}
	}
	if minElev == elevators[localElevator] {
		elevators[localElevator].Requests[order.Floor][order.Button] = config.Comfirmed
	} else {
		minElev.Requests[order.Floor][order.Button] = config.Order
	}
}

/*
	New local state stuff
*/

func comfirmOrder(elev *config.DistributorElevator) config.Request {
	for floor := range elev.Requests {
		for button := range elev.Requests[floor] {
			if elev.Requests[floor][button] == config.Order {
				elev.Requests[floor][button] = config.Comfirmed
				return config.Request{
					Floor:  floor,
					Button: config.ButtonType(button),
				}
			}
		}
	}
}

func updateLocalState(elevators []*config.DistributorElevator, elev elevator.Elevator) {
	elevators[localElevator].Behave = config.Behaviour(int(elev.Behave))
	elevators[localElevator].Floor = elev.Floor
	elevators[localElevator].Dir = config.Direction(int(elev.Dir))
}

// Todo: How to handle complete cab orders?
func checkOrderComplete(elevators []*config.DistributorElevator, elev elevator.Elevator) {
	for floor := range elev.Requests {
		for button := range elev.Requests[floor] {
			if !elev.Requests[floor][button] && elevators[localElevator].Requests[floor][button] == config.Comfirmed {
				elevators[localElevator].Requests[floor][button] = config.Complete
			}
		}
	}
}

/*
	New message from network stuff
*/
func updateElevators(elevators []*config.DistributorElevator, newElevators []config.DistributorElevator) {
	for _, newElev := range newElevators {
		elevExist := false
		for _, elev := range elevators {
			if elev.ID == newElev.ID {
				if newElev.ID == elevators[localElevator].ID {
					elev.Requests = newElev.Requests
					newOrder := comfirmOrder(elev)
				} else {
					*elev = newElev
				}
				elevExist = true
				break
			}
		}
		if !elevExist {
			tempElev := new(config.DistributorElevator)
			*tempElev = newElev
			elevators = append(elevators, tempElev)
		}
	}
}

// Må huske å skru av lysene når bestillingen er utført
func setHallLights(elevators []*config.DistributorElevator) {
	for _, elev := range elevators {
		comfirmedOrders := getComfirmedOrders(elev.Requests)
		for _, order := range comfirmedOrders {
			elevio.SetButtonLamp(elevio.ButtonType(order.Button), order.Floor, true)
		}
	}
}
