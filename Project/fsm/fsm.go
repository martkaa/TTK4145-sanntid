package fsm

import (
	"Project/elevator"
	"Project/elevio"
	"Project/request"
	"Project/timer"
	"fmt"
)

func Fsm(orderChan chan elevio.ButtonEvent, elevatorState chan<- elevator.Behaviour) {
	elev := elevator.InitElev()

	e := &elev

	elevio.Init("localhost:23456", elevator.NumFloors)

	ch_arrivedAtFloors := make(chan int)
	ch_obstr := make(chan bool)
	ch_stopButton := make(chan bool)

	ch_timer := make(chan bool)

	go elevio.PollFloorSensor(ch_arrivedAtFloors)
	go elevio.PollObstructionSwitch(ch_obstr)
	go elevio.PollStopButton(ch_stopButton)

	for {
		fmt.Println(elevator.Behaviour(e.Behave))
		elevator.LightsElev(*e)

		elevatorState <- e.Behave // Sende state til distributor, usikker på hvor ofte vi trenger å gjøre det.

		select {
		case r := <-orderChan: // Mottar ny bestilling fra distributor
			fmt.Printf("%+v\n", r)
			fsmOnRequestButtonPress(r.Floor, r.Button, e, ch_timer)

		// Alt under her er bare avhengig av heisens interne ting
		case f := <-ch_arrivedAtFloors:
			e.Floor = f
			fsmOnFloorArrival(e, ch_timer)

		case a := <-ch_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(e.Dir)
			}

		case a := <-ch_stopButton:
			fmt.Printf("%+v\n", a)
			request.RequestClearAll(e)
			e.Dir = elevio.MD_Stop
			e.Behave = elevator.Idle
			elevio.SetMotorDirection(e.Dir)
			elevio.SetDoorOpenLamp(false)
			elevator.LightsElev(*e)

		case <-ch_timer:
			e.TimerCount -= 1
			if e.TimerCount == 0 {
				fsmOnDoorTimeout(e)
			}
		}
	}
}

func fsmOnFloorArrival(e *elevator.Elevator, ch_timer chan<- bool) {

	switch {
	case e.Behave == elevator.Moving:
		if request.RequestShouldStop(e) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevator.LightsElev(*e)
			request.RequestClearAtCurrentFloor(e)
			elevio.SetDoorOpenLamp(true)
			go timer.TimerDoor(elevator.DoorOpenDuration, ch_timer, e)
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

func fsmOnRequestButtonPress(btnFloor int, btnType elevio.ButtonType, e *elevator.Elevator, ch_timer chan<- bool) {
	switch {
	case e.Behave == elevator.DoorOpen:
		if e.Floor == btnFloor {
			go timer.TimerDoor(elevator.DoorOpenDuration, ch_timer, e)
		} else {
			e.Requests[btnFloor][int(btnType)] = true
		}
	case e.Behave == elevator.Moving:
		e.Requests[btnFloor][int(btnType)] = true
	case e.Behave == elevator.Idle:
		if e.Floor == btnFloor {
			elevator.LightsElev(*e)
			elevio.SetDoorOpenLamp(true)
			go timer.TimerDoor(elevator.DoorOpenDuration, ch_timer, e)
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
