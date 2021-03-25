package fsm

import (
	"Project/Timer"
	"Project/elevator"
	"Project/elevio"
	"Project/request"
)

func FsmOnFloorArrival(e *elevator.Elevator, timerChan chan<- bool) {

	switch {
	case e.Behave == elevator.Moving:
		if request.RequestShouldStop(e) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevator.LightsElev(*e)
			request.RequestClearAtCurrentFloor(e)
			elevio.SetDoorOpenLamp(true)
			go timer.TimerDoor(elevator.DoorOpenDuration, timerChan, e)
			e.Behave = elevator.DoorOpen
		}
	default:
		break
	}
}

func FsmOnDoorTimeout(e *elevator.Elevator) {
	switch {
	case e.Behave == elevator.DoorOpen:
		request.RequestChooseDirection(e)
		elevio.SetMotorDirection(e.Dir)
		elevio.SetDoorOpenLamp(false)

		if e.Dir == elevio.MD_Stop {
			e.Behave = elevator.Idle
		} else {
			e.Behave = elevator.Moving
		}
	default:
		break
	}
}

func FsmOnRequestButtonPress(btnFloor int, btnType elevio.ButtonType, e *elevator.Elevator, timerChan chan<- bool) {
	switch {
	case e.Behave == elevator.DoorOpen:
		if e.Floor == btnFloor {
			go timer.TimerDoor(elevator.DoorOpenDuration, timerChan, e)
		} else {
			e.Requests[btnFloor][int(btnType)] = true
		}
	case e.Behave == elevator.Moving:
		e.Requests[btnFloor][int(btnType)] = true
	case e.Behave == elevator.Idle:
		if e.Floor == btnFloor {
			elevator.LightsElev(*e)
			elevio.SetDoorOpenLamp(true)
			go timer.TimerDoor(elevator.DoorOpenDuration, timerChan, e)
			e.Behave = elevator.DoorOpen
			break
		} else {
			e.Requests[btnFloor][int(btnType)] = true
			request.RequestChooseDirection(e)
			elevio.SetMotorDirection(e.Dir)
			e.Behave = elevator.Moving
			break
		}
	}
}
