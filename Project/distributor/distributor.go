package distributor

import (
	"Project/config"
	"Project/cost"
	"Project/elevator"
	"Project/elevio"
	"Project/fsm"
	"Project/network/bcast"
	"fmt"
	"time"
)

const localElevator = 0

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
	//for i := 0; i < 5; i++ {
	ch_transmit <- tempElevators
	//}
}

func printRequests(elevators []*config.DistributorElevator) {
	for _, elev := range elevators {
		fmt.Println(elev.Requests)
	}
}

func printNewRequests(elevators []config.DistributorElevator) {
	for _, elev := range elevators {
		fmt.Println(elev.Requests)
	}
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
	go bcast.Transmitter(16568, ch_msgToNetwork)
	go bcast.Receiver(16568, ch_msgFromNetwork)

	/* Initialiaze elevator */
	elevators := make([]*config.DistributorElevator, 0)
	thisElevator := new(config.DistributorElevator)
	*thisElevator = elevatorInit(id)
	elevators = append(elevators, thisElevator)

	for {
		select {
		case newOrder := <-ch_newLocalOrder:
			//fmt.Println("New Order")
			assignOrder(elevators, newOrder)
			if elevators[localElevator].Requests[newOrder.Floor][newOrder.Button] == config.Order {
				//fmt.Println("Send a 1 before sending to local elevator")
				broadcast(elevators, ch_msgToNetwork)
				time.Sleep(time.Millisecond * 100)
				elevators[localElevator].Requests[newOrder.Floor][newOrder.Button] = config.Comfirmed
				setHallLights(elevators)
				ch_orderToLocal <- newOrder
			}
			broadcast(elevators, ch_msgToNetwork)
			time.Sleep(time.Millisecond * 100)

		case newState := <-ch_newLocalState:
			//fmt.Println("New state")
			updateLocalState(elevators, newState)
			checkLocalOrderComplete(elevators[localElevator], newState)
			//fmt.Println("set hall lights msg to network")
			setHallLights(elevators)
			broadcast(elevators, ch_msgToNetwork)
			time.Sleep(time.Millisecond * 100)
			removeCompletedOrders(elevators)

		case newElevators := <-ch_msgFromNetwork:
			fmt.Println(newElevators[0].ID)
			printNewRequests(newElevators)
			//fmt.Println("------------New above, updated below-------------")
			updateElevators(elevators, newElevators)
			printRequests(elevators)
			addNewElevator(&elevators, newElevators)
			extractNewOrder := comfirmNewOrder(elevators[localElevator])
			setHallLights(elevators)
			removeCompletedOrders(elevators)
			if extractNewOrder.Floor != config.NoOrder {
				tempOrder := elevio.ButtonEvent{
					Button: elevio.ButtonType(extractNewOrder.Button),
					Floor:  extractNewOrder.Floor}
				ch_orderToLocal <- tempOrder
				broadcast(elevators, ch_msgToNetwork)
				time.Sleep(time.Millisecond * 100)
			}
		}
	}
}

/*
	New order from local stuff
*/

func assignOrder(elevators []*config.DistributorElevator, order elevio.ButtonEvent) {
	if len(elevators) < 2 {
		elevators[localElevator].Requests[order.Floor][order.Button] = config.Order
		return
	}

	if order.Button == elevio.BT_Cab {
		elevators[localElevator].Requests[order.Floor][order.Button] = config.Order
		return
	}

	minCost := 99999
	elevCost := 0
	var minElev *config.DistributorElevator

	for _, elev := range elevators {
		elevCost = cost.Cost(elev, order)
		fmt.Println(elevCost)
		if elevCost < minCost {
			minCost = elevCost
			minElev = elev
		}
	}
	if minElev.ID == elevators[localElevator].ID {
		elevators[localElevator].Requests[order.Floor][order.Button] = config.Order
	} else {
		minElev.Requests[order.Floor][order.Button] = config.Order
	}
}

/*
	New local state stuff
*/

func removeCompletedOrders(elevators []*config.DistributorElevator) {
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

func checkLocalOrderComplete(elev *config.DistributorElevator, localElev elevator.Elevator) {
	for floor := range elev.Requests {
		for button := range elev.Requests[floor] {
			if !localElev.Requests[floor][button] && elev.Requests[floor][button] == config.Comfirmed {
				elev.Requests[floor][button] = config.Complete
			}
		}
	}
}

/*
	New message from network stuff
*/

func copyRequests(elev *config.DistributorElevator, newElev config.DistributorElevator) {
	for floor := range elev.Requests {
		for button := range elev.Requests[floor] {
			elev.Requests[floor][button] = newElev.Requests[floor][button]
		}
	}
}

func updateElevators(elevators []*config.DistributorElevator, newElevators []config.DistributorElevator) {
	for _, elev := range elevators {
		if elev.ID == newElevators[localElevator].ID {
			*elev = newElevators[localElevator]
		}
	}
	for _, newElev := range newElevators {
		if newElev.ID == elevators[localElevator].ID {
			for floor := range newElev.Requests {
				for button := range newElev.Requests[floor] {
					if newElev.Requests[floor][button] == config.Order {
						elevators[localElevator].Requests[floor][button] = config.Order
					}
				}
			}
		}
	}
}

/*	if newElevators[localElevator].ID != elevators[localElevator].ID {
		for _, newElev := range newElevators {
			for _, elev := range elevators {
				isLocalElev := true
				if elev.ID == newElev.ID {
					if newElev.ID != elevators[localElevator].ID {
						isLocalElev = false
						elev.Behave = newElev.Behave
						elev.Dir = newElev.Dir
						elev.Floor = newElev.Floor
					}
					updateRequests(elev, newElev, isLocalElev)
				}
			}
		}
	}
}*/

func addNewElevator(elevators *[]*config.DistributorElevator, newElevators []config.DistributorElevator) {
	for _, newElev := range newElevators {
		elevExist := false
		for _, elev := range *elevators {
			if elev.ID == newElev.ID {
				elevExist = true
				break
			}
		}
		if !elevExist {
			tempElev := new(config.DistributorElevator)
			*tempElev = elevatorInit(newElev.ID)
			(*tempElev).Behave = newElev.Behave
			(*tempElev).Dir = newElev.Dir
			(*tempElev).Floor = newElev.Floor
			copyRequests(tempElev, newElev)
			*elevators = append(*elevators, tempElev)
		}
	}
}

func comfirmNewOrder(elev *config.DistributorElevator) config.Request {
	for floor := range elev.Requests {
		for button := 0; button < len(elev.Requests[floor])-1; button++ {
			if elev.Requests[floor][button] == config.Order {
				elev.Requests[floor][button] = config.Comfirmed
				return config.Request{
					Floor:  floor,
					Button: config.ButtonType(button)}
			}
		}
	}
	return config.Request{
		Floor:  config.NoOrder,
		Button: config.HallUp}
}

func setHallLights(elevators []*config.DistributorElevator) {
	for _, elev := range elevators {
		for floor := range elev.Requests {
			for button := 0; button < len(elev.Requests[floor])-1; button++ {
				if elev.Requests[floor][button] == config.Comfirmed {
					elevio.SetButtonLamp(elevio.ButtonType(button), floor, true)
				}
				if elev.Requests[floor][button] == config.Complete {
					elevio.SetButtonLamp(elevio.ButtonType(button), floor, false)
				}
			}
		}
	}
}
