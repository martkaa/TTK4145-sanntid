package distributor

import (
	"Project/assigner"
	"Project/config"
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/network/peers"
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
	ch_peerUpdate chan peers.PeerUpdate,
	ch_watchdogElevatorStuck chan bool,
	ch_elevStuck chan bool,
	ch_clearLocalHallOrders chan bool) {

	/* Initialiaze elevator */
	elevators := make([]*config.DistributorElevator, 0)
	thisElevator := new(config.DistributorElevator)
	*thisElevator = elevatorInit(id)
	elevators = append(elevators, thisElevator)

	connectTimer := time.NewTimer(time.Duration(3) * time.Second)
	select {
	case newElevators := <-ch_msgFromNetwork:
		for _, elev := range newElevators {
			if elev.ID == elevators[localElevator].ID {
				for floor := range elev.Requests {
					if elev.Requests[floor][config.Cab] == config.Comfirmed ||
						elev.Requests[floor][config.Cab] == config.Order {
						ch_newLocalOrder <- elevio.ButtonEvent{
							Floor:  floor,
							Button: elevio.ButtonType(int(config.Cab))}
					}
				}
			}
		}
		break
	case <-connectTimer.C:
		break
	}

	for {
		select {
		case newOrder := <-ch_newLocalOrder:
			assigner.AssignOrder(elevators, newOrder)
			if elevators[localElevator].Requests[newOrder.Floor][newOrder.Button] == config.Order {
				broadcast(elevators, ch_msgToNetwork)
				elevators[localElevator].Requests[newOrder.Floor][newOrder.Button] = config.Comfirmed
				setHallLights(elevators)
				ch_orderToLocal <- newOrder
			}
			broadcast(elevators, ch_msgToNetwork)
			setHallLights(elevators)

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

		case newElevators := <-ch_msgFromNetwork:
			updateElevators(elevators, newElevators)
			assigner.ReassignOrders(elevators, ch_newLocalOrder)
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
		case peer := <-ch_peerUpdate:
			if len(peer.Lost) != 0 {
				for _, stringLostID := range peer.Lost {
					for _, elev := range elevators {
						if stringLostID == elev.ID {
							elev.Behave = config.Unavailable
						}
					}
					assigner.ReassignOrders(elevators, ch_newLocalOrder)
				}
			}
			setHallLights(elevators)
			broadcast(elevators, ch_msgToNetwork)
		case <-ch_watchdogElevatorStuck:
			elevators[localElevator].Behave = config.Unavailable
			broadcast(elevators, ch_msgToNetwork)
			for floor := range elevators[localElevator].Requests {
				for button := 0; button < len(elevators[localElevator].Requests[floor])-1; button++ {
					elevators[localElevator].Requests[floor][button] = config.None
				}
			}
			setHallLights(elevators)
			ch_clearLocalHallOrders <- true
		}
	}
}

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
				//elevio.SetButtonLamp(elevio.ButtonType(button), floor, false)
			}
			if localElev.Requests[floor][button] && elev.Requests[floor][button] != config.Comfirmed && elev.Behave != config.Unavailable {
				elev.Requests[floor][button] = config.Comfirmed
			}
		}
	}
}

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
				for floor := range elev.Requests {
					for button := range elev.Requests[floor] {
						if !(elev.Requests[floor][button] == config.Comfirmed && newElevators[localElevator].Requests[floor][button] == config.Order) {
							elev.Requests[floor][button] = newElevators[localElevator].Requests[floor][button]
						}
						elev.Floor = newElevators[localElevator].Floor
						elev.Dir = newElevators[localElevator].Dir
						elev.Behave = newElevators[localElevator].Behave
					}
				}
			}
		}
		for _, newElev := range newElevators {
			if newElev.ID == elevators[localElevator].ID {
				if elevators[localElevator].Behave != config.Unavailable {
					for floor := range newElev.Requests {
						for button := range newElev.Requests[floor] {
							if newElev.Requests[floor][button] == config.Order {
								(*elevators[localElevator]).Requests[floor][button] = config.Order
							}
						}
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
				//elevio.SetButtonLamp(elevio.ButtonType(button), floor, true)
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
