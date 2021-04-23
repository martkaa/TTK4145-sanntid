package distributor

import (
	"Project/assigner"
	"Project/config"
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/network/peers"
	"time"
)

const localElevator = 0

// Initialize elevator states
func distributorElevatorInit(id string) config.DistributorElevator {
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

func Distributor(
	id string,
	ch_newLocalOrder chan elevio.ButtonEvent,
	ch_newLocalState chan elevator.Elevator,
	ch_msgFromNetwork chan []config.DistributorElevator,
	ch_msgToNetwork chan []config.DistributorElevator,
	ch_orderToLocal chan elevio.ButtonEvent,
	ch_peerUpdate chan peers.PeerUpdate,
	ch_petWatchdogStuck chan bool,
	ch_watchdogStuckBark chan bool,
	ch_clearLocalHallOrders chan bool) {

	elevators := make([]*config.DistributorElevator, 0)
	thisElevator := new(config.DistributorElevator)
	*thisElevator = distributorElevatorInit(id)
	elevators = append(elevators, thisElevator)

	connectTimer := time.NewTimer(time.Duration(config.ReconnectTimerSec) * time.Second)
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
				ch_petWatchdogStuck <- false
			}
			for floor := range elevators[config.LocalElevator].Requests {
				for button := range elevators[config.LocalElevator].Requests[floor] {
					if !newState.Requests[floor][button] &&
						elevators[config.LocalElevator].Requests[floor][button] == config.Comfirmed {
						elevators[config.LocalElevator].Requests[floor][button] = config.Complete
					}
					if elevators[config.LocalElevator].Behave != config.Unavailable &&
						newState.Requests[floor][button] &&
						elevators[config.LocalElevator].Requests[floor][button] != config.Comfirmed {
						elevators[config.LocalElevator].Requests[floor][button] = config.Comfirmed
					}
				}
			}
			setHallLights(elevators)
			broadcast(elevators, ch_msgToNetwork)
			removeCompletedOrders(elevators)

		case newElevators := <-ch_msgFromNetwork:
			updateElevators(elevators, newElevators)
			assigner.ReassignOrders(elevators, ch_newLocalOrder)
			for _, newElev := range newElevators {
				elevExist := false
				for _, elev := range elevators {
					if elev.ID == newElev.ID {
						elevExist = true
						break
					}
				}
				if !elevExist {
					addNewElevator(&elevators, newElev)
				}
			}
			extractNewOrder := comfirmNewOrder(elevators[localElevator])
			setHallLights(elevators)
			removeCompletedOrders(elevators)
			if extractNewOrder != nil {
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
						assigner.ReassignOrders(elevators, ch_newLocalOrder)
						for floor := range elev.Requests {
							for button := 0; button < len(elev.Requests[floor])-1; button++ {
								elev.Requests[floor][button] = config.None
							}
						}
					}
				}
			}
			setHallLights(elevators)
			broadcast(elevators, ch_msgToNetwork)
		case <-ch_watchdogStuckBark:
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

func checkLocalOrderComplete(elev *config.DistributorElevator, localElev elevator.Elevator) {
	for floor := range elev.Requests {
		for button := range elev.Requests[floor] {
			if !localElev.Requests[floor][button] && elev.Requests[floor][button] == config.Comfirmed {
				elev.Requests[floor][button] = config.Complete
			}
			if localElev.Requests[floor][button] && elev.Requests[floor][button] != config.Comfirmed && elev.Behave != config.Unavailable {
				elev.Requests[floor][button] = config.Comfirmed
			}
		}
	}
}

// Updates local elevator array from received elevator array from network
func updateElevators(elevators []*config.DistributorElevator, newElevators []config.DistributorElevator) {
	if elevators[localElevator].ID != newElevators[localElevator].ID {
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
				for floor := range newElev.Requests {
					for button := range newElev.Requests[floor] {
						if elevators[localElevator].Behave != config.Unavailable {
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

// Adds newElevator to local elevator array
func addNewElevator(elevators *[]*config.DistributorElevator, newElevator config.DistributorElevator) {
	tempElev := new(config.DistributorElevator)
	*tempElev = distributorElevatorInit(newElevator.ID)
	(*tempElev).Behave = newElevator.Behave
	(*tempElev).Dir = newElevator.Dir
	(*tempElev).Floor = newElevator.Floor
	for floor := range tempElev.Requests {
		for button := range tempElev.Requests[floor] {
			tempElev.Requests[floor][button] = newElevator.Requests[floor][button]
		}
	}
	*elevators = append(*elevators, tempElev)
}

func comfirmNewOrder(elev *config.DistributorElevator) *config.Request {
	for floor := range elev.Requests {
		for button := 0; button < len(elev.Requests[floor]); button++ {
			if elev.Requests[floor][button] == config.Order {
				elev.Requests[floor][button] = config.Comfirmed
				tempOrder := new(config.Request)
				*tempOrder = config.Request{
					Floor:  floor,
					Button: config.ButtonType(button)}
				return tempOrder
			}
		}
	}
	return nil
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
