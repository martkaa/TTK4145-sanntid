package cost

import (
	"Project/config"
	"Project/elevator"
	"Project/elevio"
)

const TRAVEL_TIME = 10

const NumElevators = 4

func Cost(elev *config.DistributorElevator, req elevio.ButtonEvent) int {
	e := new(config.DistributorElevator)
	*e = *elev
	e.Requests[req.Floor][req.Button] = config.Comfirmed

	duration := 0

	switch e.Behave {
	case config.Idle:
		requestChooseDirection(e)
		if e.Dir == config.Stop {
			return duration
		}
	case config.Moving:
		duration += TRAVEL_TIME / 2
		e.Floor += int(e.Dir)
	case config.DoorOpen:
		duration -= elevator.DoorOpenDuration / 2
	}

	for {
		if requestShouldStop(*e) {
			requestClearAtCurrentFloor(e)
			duration += elevator.DoorOpenDuration
			requestChooseDirection(e)
			if e.Dir == config.Stop {
				return duration
			}
		}
		e.Floor += int(e.Dir)
		duration += TRAVEL_TIME
	}

}

func requestsAbove(elev config.DistributorElevator) bool {
	for f := elev.Floor + 1; f < config.NumFloors; f++ {
		for btn := range elev.Requests[f] {
			if elev.Requests[f][btn] == config.Comfirmed {
				return true
			}
		}
	}
	return false
}

func requestsBelow(elev config.DistributorElevator) bool {
	for f := 0; f < elev.Floor; f++ {
		for btn := range elev.Requests[f] {
			if elev.Requests[f][btn] == config.Comfirmed {
				return true
			}
		}
	}
	return false
}

func requestClearAtCurrentFloor(elev *config.DistributorElevator) {
	elev.Requests[elev.Floor][int(elevio.BT_Cab)] = config.None
	switch {
	case elev.Dir == config.Up:
		elev.Requests[elev.Floor][int(elevio.BT_HallUp)] = config.None
		if !requestsAbove(*elev) {
			elev.Requests[elev.Floor][int(elevio.BT_HallDown)] = config.None
		}
	case elev.Dir == config.Down:
		elev.Requests[elev.Floor][int(elevio.BT_HallDown)] = config.None
		if !requestsBelow(*elev) {
			elev.Requests[elev.Floor][int(elevio.BT_HallUp)] = config.None
		}
	}
}

func requestShouldStop(elev config.DistributorElevator) bool {
	switch {
	case elev.Dir == config.Down:
		return elev.Requests[elev.Floor][int(elevio.BT_HallDown)] == config.Comfirmed ||
			elev.Requests[elev.Floor][int(elevio.BT_Cab)] == config.Comfirmed ||
			!requestsBelow(elev)
	case elev.Dir == config.Up:
		return elev.Requests[elev.Floor][int(elevio.BT_HallUp)] == config.Comfirmed ||
			elev.Requests[elev.Floor][int(elevio.BT_Cab)] == config.Comfirmed ||
			!requestsAbove(elev)
	default:
		return true
	}
}

func requestChooseDirection(elev *config.DistributorElevator) {
	switch elev.Dir {
	case config.Up:
		if requestsAbove(*elev) {
			elev.Dir = config.Up
		} else if requestsBelow(*elev) {
			elev.Dir = config.Down
		} else {
			elev.Dir = config.Stop
		}
	case config.Down:
		fallthrough
	case config.Stop:
		if requestsBelow(*elev) {
			elev.Dir = config.Down
		} else if requestsAbove(*elev) {
			elev.Dir = config.Up
		} else {
			elev.Dir = config.Stop
		}
	}
}
