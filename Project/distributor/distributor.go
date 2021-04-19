package distributor

import (
	"Project/config"
	"Project/cost"
	"Project/elevator"
	"Project/elevio"
	"Project/network/bcast"
	"fmt"

	"Project/fsm"
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

func DistributorFsm(id string) {

	/* Channels */
	ch_newLocalOrder := make(chan elevio.ButtonEvent)
	ch_newLocalState := make(chan elevator.Elevator)
	ch_msgFromNetwork := make(chan []config.DistributorElevator)
	ch_msgToNetwork := make(chan []config.DistributorElevator)
	ch_orderToLocal := make(chan elevio.ButtonEvent)

	go fsm.Fsm(ch_orderToLocal, ch_newLocalState)
	go elevio.PollButtons(ch_newLocalOrder)

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
			//fmt.Println("New order")
			assignOrder(elevators, newOrder)
			//fmt.Println("New order after assign")
			if elevators[localElevator].Requests[newOrder.Floor][newOrder.Button] == config.Comfirmed {
				setHallLights(elevators)
				ch_orderToLocal <- newOrder
				fmt.Println("New order send to local")
			}
			fmt.Println("New order, before broadcast")
			broadcast(elevators, ch_msgToNetwork)

		case newState := <-ch_newLocalState:
			//fmt.Println("New State")
			updateLocalState(elevators, newState)
			//fmt.Println("New State, after updateLcalState")
			checkLocalOrderComplete(elevators, newState)
			//fmt.Println("New State, after checkOrderComplete")
			broadcast(elevators, ch_msgToNetwork)
			//fmt.Println("New State, after broadcast")
			setHallLights(elevators)
			//fmt.Println("New State, after setLights")
			/* Broadcast for some seconds */
			checkGlobalOrderComplete(elevators)
			//fmt.Println("New State, after checkGlobalOrderComplete")

		case newElevators := <-ch_msgFromNetwork:
			//fmt.Println("New Network message")
			updateElevators(elevators, newElevators)
			//fmt.Println("New Network message after update ")
			var extractNewOrder *config.Request
			comfirmNewOrder(elevators[localElevator], extractNewOrder)
			//fmt.Println("New Network message after comfirmed order")
			setHallLights(elevators)
			//fmt.Println("New Network message after set lights")
			checkGlobalOrderComplete(elevators)
			//fmt.Println("New Network message after after order complete")
			if extractNewOrder != nil {
				ch_newLocalOrder <- elevio.ButtonEvent{
					Button: elevio.ButtonType(extractNewOrder.Button),
					Floor:  extractNewOrder.Floor}
				broadcast(elevators, ch_msgToNetwork)
			}
		}
		// Error case: lost connection to elevator,
	}
}

/*
	New order from local stuff
*/

func assignOrder(elevators []*config.DistributorElevator, order elevio.ButtonEvent) {
	if len(elevators) < 2 {
		elevators[localElevator].Requests[order.Floor][order.Button] = config.Comfirmed
		return
	}

	if order.Button == elevio.BT_Cab {
		elevators[localElevator].Requests[order.Floor][order.Button] = config.Comfirmed
		return
	}

	minCost := 99999
	elevCost := 0
	var minElev *config.DistributorElevator

	for _, elev := range elevators {
		fmt.Println("Now, lets calculate the cost!")
		elevCost = cost.Cost(elev, order)
		fmt.Println("Cost calculated.")
		if elevCost < minCost {
			minCost = elevCost
			minElev = elev
		}
	}
	fmt.Println("The cost is ", minCost)
	if minElev == elevators[localElevator] {
		elevators[localElevator].Requests[order.Floor][order.Button] = config.Comfirmed
	} else {
		minElev.Requests[order.Floor][order.Button] = config.Order
	}
}

/*
	New local state stuff
*/

func checkGlobalOrderComplete(elevators []*config.DistributorElevator) {
	for _, elev := range elevators {
		for floor := range elev.Requests {
			for button := range elev.Requests[floor] {
				if elev.Requests[floor][button] == config.Complete {
					elev.Requests[floor][button] = config.None
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
func checkLocalOrderComplete(elevators []*config.DistributorElevator, elev elevator.Elevator) {
	for floor := range elev.Requests {
		for button := range elev.Requests[floor] {
			if !elev.Requests[floor][button] && elevators[localElevator].Requests[floor][button] == config.Comfirmed {
				elevators[localElevator].Requests[floor][button] = config.Complete
			}
		}
	}
}

func updateRequests(elev *config.DistributorElevator, newElev config.DistributorElevator) {
	for floor := range elev.Requests {
		for button := 1; button < len(elev.Requests[floor]); button++ {
			if int(elev.Requests[floor][button]) < int(newElev.Requests[floor][button]) {
				elev.Requests[floor][button] = newElev.Requests[floor][button]
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
				updateRequests(elev, newElev)
				if newElev.ID != elevators[localElevator].ID {
					elev.Behave = newElev.Behave
					elev.Dir = newElev.Dir
					elev.Floor = newElev.Floor
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

func comfirmNewOrder(elev *config.DistributorElevator, extractNewOrder *config.Request) {
	for floor := range elev.Requests {
		for button := range elev.Requests[floor] {
			if elev.Requests[floor][button] == config.Order {
				extractNewOrder.Floor = floor
				extractNewOrder.Button = config.ButtonType(button)
				elev.Requests[floor][button] = config.Comfirmed
			}
		}
	}
	extractNewOrder = nil
}

// Må huske å skru av lysene når bestillingen er utført
func setHallLights(elevators []*config.DistributorElevator) {
	for _, elev := range elevators {
		for floor := range elev.Requests {
			for button := range elev.Requests[floor] {
				if elev.Requests[floor][button] == config.Comfirmed {
					elevio.SetButtonLamp(elevio.ButtonType(button), floor, true)
				} else if elev.Requests[floor][button] == config.Complete {
					elevio.SetButtonLamp(elevio.ButtonType(button), floor, false)
				}
			}
		}
	}
}
