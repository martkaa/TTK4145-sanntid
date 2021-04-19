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
	DoorOpen Behaviour = 1
	Moving   Behaviour = 2
)

type Elevator struct {
	Floor      int
	Dir        elevio.MotorDirection
	Requests   [][]bool
	Behave     Behaviour
	TimerCount int
}

func InitElev() Elevator {
	requests := make([][]bool, 0)
	for floor := 0; floor < NumFloors; floor++ {
		requests = append(requests, make([]bool, NumButtons))
		for button := range requests[floor] {
			requests[floor][button] = false
		}
	}
	return Elevator{
		Floor:      0,
		Dir:        elevio.MD_Stop,
		Requests:   requests,
		Behave:     Idle,
		TimerCount: 0}
}

func LightsElev(e Elevator) {
	elevio.SetFloorIndicator(e.Floor)
	for f := range e.Requests {
		for r := range e.Requests[f] {
			elevio.SetButtonLamp(elevio.ButtonType(r), f, e.Requests[f][r])
		}
	}
}
