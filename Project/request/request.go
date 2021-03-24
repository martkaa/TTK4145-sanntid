package request

import (
	"elevator"
	"elevio"
)

type Order int

const (
	UP     Order = 0
	DOWN         = 1
	INSIDE       = 2
)

func requestsAbove(e elevator.Elevator) bool {
	for f := range e.requests {
		for btn := range e.requests[f] {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestClearAtCurrentFloor(floor int, e &elevator.Elevator) {
	e.request[floor][INSIDE]
	switch elevio.MotorDirection {
	case elevio.MD_Up:
		e.requests[floor][UP] = false
	case elevio.MD_Down:
		e.request[floor][DOWN] = false
	}
}
