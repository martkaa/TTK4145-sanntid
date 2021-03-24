package elevator

import (
	"Project/elevio"
)

const NumFloors = 4

const NumButtons = 3

const DoorOpenDuration = 3

type Behaviour int

const (
	Idle     Behaviour = 0
	DoorOpen           = 1
	Moving             = 2
)

type Elevator struct {
	Floor    int
	Dir      elevio.MotorDirection
	Requests [][]bool
	Behave   Behaviour
}

func InitElev(numFloors int, numButtons int) Elevator {
	elev := Elevator{Floor: 0, Dir: elevio.MD_Up, Requests: make([][]bool, numFloors), Behave: Moving}

	for r := range elev.Requests {
		elev.Requests[r] = make([]bool, numButtons)
	}
	return elev
}

func LightsElev(e Elevator) {
	elevio.SetFloorIndicator(e.Floor)
	for f := range e.Requests {
		for r := range e.Requests[f] {
			elevio.SetButtonLamp(elevio.ButtonType(r), f, e.Requests[f][r])
		}
	}
}
