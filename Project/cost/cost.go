package cost

import (
	"Project/config"
	"Project/elevator"
	"Project/elevio"
	"Project/request"
)

const TRAVEL_TIME = 10

const NumElevators = 4

// Kan ikke inkludere distributor her grunnet import cylcle-opplegg. Må benytte elevator-struktet
// og konvertere i distributor-moduelen.

// Struct that contains all information neccecsary to determine the elvator with the lowest cost.

func Cost(elevators []*config.DistributorElevator, req elevio.ButtonEvent, ch_assignedDistributorOrder chan config.CostRequest) {

	minElev := elevators[0]
	minCost := 999999

	for _, e := range elevators {
		elevator := DistributorElevatorToElevator(*e)
		elevCost := TimeToServeRequest(&elevator, req)
		if elevCost < minCost {
			minElev = e
			minCost = elevCost
		}
	}
	ch_assignedDistributorOrder <- config.CostRequest{
		Id:   minElev.Id,
		Cost: minCost,
		Req:  config.Request{Floor: req.Floor, Button: config.ButtonType(int(req.Button))},
	}
}

func TimeToServeRequest(e_old *elevator.Elevator, req elevio.ButtonEvent) int {
	e := e_old
	e.Requests[req.Floor][req.Button] = true

	arrivedAtRequest := false

	duration := 0

	switch e.Behave {
	case elevator.Idle:
		request.RequestChooseDirection(e)
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

/* Converts type DistributorElevator to elevator.Elevator*/
func DistributorElevatorToElevator(distElevator config.DistributorElevator) elevator.Elevator {
	req := make([][]bool, 0)
	for floor := range distElevator.Requests {
		req = append(req, make([]bool, 0))
		for button := range distElevator.Requests[floor] {
			if distElevator.Requests[floor][button] == config.Comfirmed {
				req[floor] = append(req[floor], true)
			} else {
				req[floor] = append(req[floor], false)
			}
		}
	}
	return elevator.Elevator{
		Id:         distElevator.Id,
		Floor:      distElevator.Floor,
		Dir:        elevio.MotorDirection(int(distElevator.Dir)),
		Requests:   req,
		Behave:     elevator.Behaviour(int(distElevator.Behave)),
		TimerCount: 0}
}
