package cost

import (
	"Project/distributor"
	"Project/elevator"
	"Project/elevio"
)

const NumElevators = 4

// Returnerer heis med lavest kost basert på et straffesystem/poeng
// Må deretter sorteres og delegeres

func costCalculator(request distributor.Request, elevList [NumElevators]elevator.Elevator, id int, onlineList [NumElevators]bool) int {
	if request.Btn == elevio.BT_Cab {
		return id
	}
	minCost := (elevator.NumButtons * elevator.NumFloors) * NumElevators
	bestElevator := id
	for e := 0; e < NumElevators; e++ {
		if !onlineList[e] {
			// Neglect offline elevators
			continue
		}
		cost := request.Floor - elevList[e].Floor

		if cost == 0 && elevList[e].Behave != elevator.Moving {
			bestElevator = e
			return bestElevator
		}
		if cost < 0 {
			cost = -cost
			if elevList[e].Dir == elevio.MD_Up {
				cost += 3
			}

		} else if cost > 0 {
			if elevList[e].Dir == elevio.MD_Down {
				cost += 3
			}
		}
		if cost == 0 && elevList[e].Behave == elevator.Moving {
			cost += 4
		}
		if elevList[e].Behave == elevator.DoorOpen {
			cost++
		}
		if cost < minCost {
			minCost = cost
			bestElevator = e
		}
	}
	return bestElevator
}

func distrubute() {
	var (
		elevList       [NumElevators]elevator.Elevator
		onlineList     [NumElevators]bool
		completedOrder Keypress
	)
	completedOrder.DesignatedElevator = id
	elevList[id] = <-elevatorCh
	updateSyncCh <- elevList
}
