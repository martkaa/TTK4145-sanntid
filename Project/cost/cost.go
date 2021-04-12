package cost

import (
	"Project/elevType"
	"Project/elevator"
	"Project/elevio"
	"Project/request"
)

const TRAVEL_TIME = 10

const NumElevators = 4

// Kan ikke inkludere distributor her grunnet import cylcle-opplegg. Må benytte elevator-struktet
// og konvertere i distributor-moduelen.

// Struct that contains all information neccecsary to determine the elvator with the lowest cost.

func Cost(elevators []*elevType.Distributor, req elevio.ButtonEvent, ch_assignedDistributorOrder chan *elevType.Distributor) {

	minElev := elevators[0]
	minCost := 999999

	for _, e := range elevators {
		elevCost := TimeToServeRequest(e, req)
		if elevCost < minCost {
			minElev = e
			minCost = elevCost
		}
	}
	ch_assignedDistributorOrder <- minElev
}

func TimeToServeRequest(e_old *elevType.Distributor, req elevio.ButtonEvent) int {
	e := e_old
	e.Requests[req.Floor][req.Button] = true

	arrivedAtRequest := false

	duration := 0

	switch e.Behave {
	case elevType.Idle:
		request.RequestChooseDirection(e)
		if e.Dir == elevType.Stop {
			return duration
		}
	case elevType.Moving:
		duration += TRAVEL_TIME / 2
		e.Floor += int(e.Dir)
	case elevType.DoorOpen:
		duration -= elevator.DoorOpenDuration / 2
	}

	for {
		if request.RequestShouldStop(e) {
			request.RequestClearAtCurrentFloor(e)
			if arrivedAtRequest {
				return duration
			}
			duration += elevator.DoorOpenDuration
			request.RequestChooseDirection(e)
		}
		e.Floor += int(e.Dir)
		duration += TRAVEL_TIME
	}

}
