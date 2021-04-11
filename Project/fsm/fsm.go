package fsm

import (
	"Project/elevator"
	"Project/elevio"
	"Project/request"
	"Project/timer"
	"fmt"
)

func Fsm(orderChan chan elevio.ButtonEvent, elevatorState chan<- elevator.Behaviour) {
	elev := elevator.InitElev(elevator.NumFloors, elevator.NumButtons)

	e := &elev

	elevio.Init("localhost:23456", elevator.NumFloors)

	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	timerChan := make(chan bool)

	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	for {
		fmt.Println(elevator.Behaviour(e.Behave))
		elevator.LightsElev(*e)

		elevatorState <- e.Behave // Sende state til distributor, usikker på hvor ofte vi trenger å gjøre det.

		select {
		case r := <-orderChan: // Mottar ny bestilling fra distributor
			fmt.Printf("%+v\n", r)
			fsmOnRequestButtonPress(r.Floor, r.Button, e, timerChan)

		// Alt under her er bare avhengig av heisens interne ting
		case f := <-drv_floors:
			e.Floor = f
			fsmOnFloorArrival(e, timerChan)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(e.Dir)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			request.RequestClearAll(e)
			e.Dir = elevio.MD_Stop
			e.Behave = elevator.Idle
			elevio.SetMotorDirection(e.Dir)
			elevio.SetDoorOpenLamp(false)
			elevator.LightsElev(*e)

		case <-timerChan:
			e.TimerCount -= 1
			if e.TimerCount == 0 {
				fsmOnDoorTimeout(e)
			}
		}
	}
}

func fsmOnFloorArrival(e *elevator.Elevator, timerChan chan<- bool) {

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

func fsmOnDoorTimeout(e *elevator.Elevator) {
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

func fsmOnRequestButtonPress(btnFloor int, btnType elevio.ButtonType, e *elevator.Elevator, timerChan chan<- bool) {
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
