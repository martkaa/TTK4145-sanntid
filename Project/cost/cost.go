package cost

import (
	"Project/elevator"
	"Project/elevio"
	"Project/request"
)

const TRAVEL_TIME = 10

const NumElevators = 4

// Kan ikke inkludere distributor her grunnet import cylcle-opplegg. Må benytte elevator-struktet
// og konvertere i distributor-moduelen.

// Struct that contains all information neccecsary to determine the elvator with the lowest cost.
type CostElevator struct {
	Id   int
	Elev elevator.Elevator
	Req  elevio.ButtonEvent
}

func Cost(elevators []CostElevator, req elevio.ButtonEvent, ch_assignedDistributorOrder chan CostElevator) {

	minElev := elevators[0]
	minCost := 999999

	for _, e := range elevators {
		elevCost := TimeToServeRequest(e, req)
		if elevCost < minCost {
			minElev = e
			minCost = elevCost
		}
	}
	ch_assignedDistributorOrder <- CostElevator{Id: minElev.Id, Req: req}
}

func TimeToServeRequest(e_old CostElevator, req elevio.ButtonEvent) int {
	e := e_old
	e.Elev.Requests[req.Floor][req.Button] = true

	arrivedAtRequest := false

	duration := 0

	switch e.Elev.Behave {
	case elevator.Idle:
		request.RequestChooseDirection(&e.Elev)
		if e.Elev.Dir == elevio.MD_Stop {
			return duration
		}
	case elevator.Moving:
		duration += TRAVEL_TIME / 2
		e.Elev.Floor += int(e.Elev.Dir)
	case elevator.DoorOpen:
		duration -= elevator.DoorOpenDuration / 2
	}

	for {
		if request.RequestShouldStop(&e.Elev) {
			request.RequestClearAtCurrentFloor(&e.Elev)
			if arrivedAtRequest {
				return duration
			}
			duration += elevator.DoorOpenDuration
			request.RequestChooseDirection(&e.Elev)
		}
		e.Elev.Floor += int(e.Elev.Dir)
		duration += TRAVEL_TIME
	}

}
