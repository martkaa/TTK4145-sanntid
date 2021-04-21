package distributor

import (
	"Project/config"
	"Project/cost"
	"Project/elevator"
	"Project/elevio"
	"fmt"
	"strconv"
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
	ch_transmit <- tempElevators
	time.Sleep(time.Millisecond * 50)
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

func DistributorFsm(
	id string,
	ch_newLocalOrder chan elevio.ButtonEvent,
	ch_newLocalState chan elevator.Elevator,
	ch_msgFromNetwork chan []config.DistributorElevator,
	ch_msgToNetwork chan []config.DistributorElevator,
	ch_orderToLocal chan elevio.ButtonEvent,
	ch_watchdogLostConnection chan string,
	ch_watchdogElevatorStuck chan bool,
	ch_elevStuck chan bool) {

	/* Initialiaze elevator */
	elevators := make([]*config.DistributorElevator, 0)
	thisElevator := new(config.DistributorElevator)
	*thisElevator = elevatorInit(id)
	elevators = append(elevators, thisElevator)

	for {
		select {
		case newOrder := <-ch_newLocalOrder:
			assignOrder(elevators, newOrder)
			if elevators[localElevator].Requests[newOrder.Floor][newOrder.Button] == config.Order {
				broadcast(elevators, ch_msgToNetwork)
				elevators[localElevator].Requests[newOrder.Floor][newOrder.Button] = config.Comfirmed
				setHallLights(elevators)
				ch_orderToLocal <- newOrder
			}
			broadcast(elevators, ch_msgToNetwork)
			time.Sleep(time.Microsecond * 50)

		case newState := <-ch_newLocalState:
			if newState.Floor != elevators[localElevator].Floor ||
				newState.Behave == elevator.Idle ||
				newState.Behave == elevator.DoorOpen {
				elevators[localElevator].Behave = config.Behaviour(int(newState.Behave))
				elevators[localElevator].Floor = newState.Floor
				elevators[localElevator].Dir = config.Direction(int(newState.Dir))
				ch_elevStuck <- false
			} else {
				ch_elevStuck <- true
			}
			checkLocalOrderComplete(elevators[localElevator], newState)
			setHallLights(elevators)
			broadcast(elevators, ch_msgToNetwork)
			removeCompletedOrders(elevators)
			time.Sleep(time.Microsecond * 50)

		case newElevators := <-ch_msgFromNetwork:
			updateElevators(elevators, newElevators)
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
			}
			time.Sleep(time.Microsecond * 50)

		case lostID := <-ch_watchdogLostConnection:
			reassignOrders(elevators, lostID, ch_orderToLocal)
			//Delete elevator from elevators

		case <-ch_watchdogElevatorStuck:
			reassignOrders(elevators, elevators[localElevator].ID, ch_orderToLocal)
			elevators[localElevator].Behave = config.MotorStop
		}
	}
}

/*
	New order from local stuff
*/

func assignOrder(elevators []*config.DistributorElevator, order elevio.ButtonEvent) {
	if len(elevators) < 2 || order.Button == elevio.BT_Cab {
		elevators[localElevator].Requests[order.Floor][order.Button] = config.Order
		return
	}
	minCost := 99999
	elevCost := 0
	var minElev *config.DistributorElevator
	for _, elev := range elevators {
		if elev.Behave != config.MotorStop {
			elevCost = cost.Cost(elev, order)
			if elevCost < minCost {
				minCost = elevCost
				minElev = elev
			}
		}
	}
	minElev.Requests[order.Floor][order.Button] = config.Order
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
	if elevators[0].ID != newElevators[0].ID {
		for _, elev := range elevators {
			if elev.ID == newElevators[localElevator].ID {
				*elev = newElevators[localElevator]
			}
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
		for button := 0; button < len(elev.Requests[floor]); button++ {
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
	for button := 0; button < config.NumButtons-1; button++ {
		for floor := 0; floor < config.NumFloors; floor++ {
			isLight := false
			for _, elev := range elevators {
				if elev.Requests[floor][button] == config.Comfirmed {
					isLight = true
				}
			}
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, isLight)
		}
	}
}

func reassignOrders(elevators []*config.DistributorElevator, errorID string, ch_orderToLocal chan elevio.ButtonEvent) {
	var lostElev *config.DistributorElevator
	if errorID != elevators[localElevator].ID {
		lowestID := 999
		for _, elev := range elevators {
			ID, _ := strconv.Atoi(elev.ID)
			if ID < lowestID && elev.ID != errorID {
				lowestID = ID
			}
			if elev.ID == errorID {
				lostElev = elev
			}
		}
		if elevators[0].ID == strconv.Itoa(lowestID) {
			for floor := range lostElev.Requests {
				for button := 0; button < len(lostElev.Requests[floor])-1; button++ {
					if lostElev.Requests[floor][button] == config.Order ||
						lostElev.Requests[floor][button] == config.Comfirmed {
						ch_orderToLocal <- elevio.ButtonEvent{
							Floor:  floor,
							Button: elevio.ButtonType(button)}
					}
				}
			}
		}
	}
	for floor := range lostElev.Requests {
		for button := 0; button < len(lostElev.Requests[floor])-1; button++ {
			lostElev.Requests[floor][button] = config.None
		}
	}
}
