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
	minElev.Elev.Requests[req.Floor][req.Button] = true
	ch_assignedDistributorOrder <- CostElevator{Elev: minElev, Req: req}
}

func TimeToServeRequest(e_old elevator.Elevator, req elevio.ButtonEvent) int {
	e := e_old
	e.Requests[req.Floor][req.Button] = true

	arrivedAtRequest := false

	duration := 0

	switch e.Behave {
	case elevator.Idle:
		request.RequestChooseDirection(&e)
		if e.Dir == elevio.MD_Stop {
			return duration
		}
	case elevator.Moving:
		duration += TRAVEL_TIME / 2
		e.Floor += int(e.Dir)
	case elevator.DoorOpen:
		duration -= elevator.DoorOpenDuration / 2
	}

	for {
		if request.RequestShouldStop(&e) {
			request.RequestClearAtCurrentFloor(&e)
			if arrivedAtRequest {
				return duration
			}
			duration += elevator.DoorOpenDuration
			request.RequestChooseDirection(&e)
		}
		e.Floor += int(e.Dir)
		duration += TRAVEL_TIME
	}

}
