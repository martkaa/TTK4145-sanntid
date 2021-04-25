package assigner

import (
	"Project/assigner/cost"
	"Project/config"
	"Project/localElevator/elevio"
	"strconv"
)

// Checks for unavailable elevators and reassigns eventual orders if an elevator is unavailable.
func ReassignOrders(elevators []*config.DistributorElevator, ch_newLocalOrder chan elevio.ButtonEvent) {
	lowestID := 999
	for _, elev := range elevators {
		if elev.Behave != config.Unavailable {
			ID, _ := strconv.Atoi(elev.ID)
			if ID < lowestID {
				lowestID = ID
			}
		}
	}
	for _, elev := range elevators {
		if elev.Behave == config.Unavailable {
			for floor := range elev.Requests {
				for button := 0; button < len(elev.Requests[floor])-1; button++ {
					if elev.Requests[floor][button] == config.Order ||
						elev.Requests[floor][button] == config.Comfirmed {
						if elevators[config.LocalElevator].ID == strconv.Itoa(lowestID) {
							ch_newLocalOrder <- elevio.ButtonEvent{
								Floor:  floor,
								Button: elevio.ButtonType(button)}
						}
					}
				}
			}
		}
	}
}

// Assignes new order to the right elevator depending on a cost function.
func AssignOrder(elevators []*config.DistributorElevator, order elevio.ButtonEvent) {
	if len(elevators) < 2 || order.Button == elevio.BT_Cab {
		elevators[config.LocalElevator].Requests[order.Floor][order.Button] = config.Order
		return
	}
	minCost := 99999
	elevCost := 0
	var minElev *config.DistributorElevator
	for _, elev := range elevators {
		elevCost = cost.Cost(elev, order)
		if elevCost < minCost {
			minCost = elevCost
			minElev = elev
		}
	}
	(*minElev).Requests[order.Floor][order.Button] = config.Order
}
