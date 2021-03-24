package elevator

import (
	"elevio"
)

type Behaviour int

const (
	Idle     Behaviour = 0
	DoorOpen           = 1
	Moving             = 2
)

type Elevator struct {
	floor     int
	dir       elevio.MotorDirection
	requests  [][]bool
	behaviour Behaviour
}

func InitElev(numFloors int, numButtons int) Elevator {
	elev := Elevator{floor: 0, dir: elevio.MD_Stop, requests: make([][]bool, numFloors), behaviour: Idle}

	for r := range elev.requests {
		elev.requests[r] = make([]bool, numButtons)
	}

	return elev
}
