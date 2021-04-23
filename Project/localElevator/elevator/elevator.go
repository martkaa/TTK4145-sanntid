package elevator

import (
	"Project/config"
	"Project/localElevator/elevio"
)

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

// Initialize the local elevator without orders and in floor 0 in IDLE state.
func InitElev() Elevator {
	requests := make([][]bool, 0)
	for floor := 0; floor < config.NumFloors; floor++ {
		requests = append(requests, make([]bool, config.NumButtons))
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

// Checks for cab orders and updates the lights accordingly.
func LightsElev(e Elevator) {
	elevio.SetFloorIndicator(e.Floor)
	for f := range e.Requests {
		elevio.SetButtonLamp(elevio.ButtonType(elevio.BT_Cab), f, e.Requests[f][elevio.BT_Cab])
	}
}
