package fsm

import (
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/request"
	"Project/localElevator/timer"
	"time"
)

func Fsm(ch_orderChan chan elevio.ButtonEvent, ch_elevatorState chan<- elevator.Elevator, ch_clearLocalHallOrders chan bool) {
	elev := elevator.InitElev()

	e := &elev

	elevio.SetDoorOpenLamp(false)

	ch_arrivedAtFloors := make(chan int)
	ch_obstruction := make(chan bool)
	ch_stopButton := make(chan bool)

	ch_timerDoor := make(chan bool)
	ch_timerUpdateState := make(chan bool)

	go elevio.PollFloorSensor(ch_arrivedAtFloors)
	go elevio.PollObstructionSwitch(ch_obstruction)
	go elevio.PollStopButton(ch_stopButton)

	go timer.TimerUpdateState(500, ch_timerUpdateState)

	elevio.SetMotorDirection(elevio.MD_Down)

	for {
		floor := <-ch_arrivedAtFloors
		if floor != 0 {
			elevio.SetMotorDirection(elevio.MD_Down)
		} else {
			elevio.SetMotorDirection(elevio.MD_Stop)
			break
		}
	}
	ch_elevatorState <- *e

	doorTimer := time.NewTimer(time.Duration(elevator.DoorOpenDuration) * time.Second)
	doorTimer.Stop()

	for {
		elevator.LightsElev(*e)

		select {
		case r := <-ch_orderChan: // Mottar ny bestilling fra distributor
			fsmOnRequestButtonPress(r.Floor, r.Button, e, ch_timerDoor, ch_elevatorState)

		// Alt under her er bare avhengig av heisens interne ting
		case f := <-ch_arrivedAtFloors:
			e.Floor = f
			fsmOnFloorArrival(e, ch_timerDoor, ch_elevatorState, doorTimer)
		case <-doorTimer.C:
			fsmOnDoorTimeout(e, ch_elevatorState)
		case <-ch_timerUpdateState:
			ch_elevatorState <- *e
			time.Sleep(time.Millisecond * 50)
		case <-ch_clearLocalHallOrders:
			request.RequestClearHall(e)
		case obstruction := <-ch_obstruction:
			if e.Behave == elevator.DoorOpen && obstruction {
				doorTimer.Reset(time.Duration(elevator.DoorOpenDuration) * time.Second)
			}
		}

	}
}

func fsmOnFloorArrival(e *elevator.Elevator, ch_timer chan<- bool, ch_elevatorState chan<- elevator.Elevator, doorTimer *time.Timer) {

	switch {
	case e.Behave == elevator.Moving:
		if request.RequestShouldStop(e) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevator.LightsElev(*e)
			request.RequestClearAtCurrentFloor(e)
			elevio.SetDoorOpenLamp(true)
			doorTimer.Reset(time.Duration(elevator.DoorOpenDuration) * time.Second)
			e.Behave = elevator.DoorOpen
			ch_elevatorState <- *e
			time.Sleep(time.Millisecond * 50)
		}
	default:
		break
	}
}

func fsmOnDoorTimeout(e *elevator.Elevator, ch_elevatorState chan<- elevator.Elevator) {
	switch {
	case e.Behave == elevator.DoorOpen:
		request.RequestChooseDirection(e)
		elevio.SetMotorDirection(e.Dir)
		elevio.SetDoorOpenLamp(false)

		if e.Dir == elevio.MD_Stop {
			e.Behave = elevator.Idle
			ch_elevatorState <- *e
			time.Sleep(time.Millisecond * 50)
		} else {
			e.Behave = elevator.Moving
			ch_elevatorState <- *e
			time.Sleep(time.Millisecond * 50)
		}
	default:
		break
	}
}

func fsmOnRequestButtonPress(btnFloor int, btnType elevio.ButtonType, e *elevator.Elevator, ch_timer chan<- bool, ch_elevatorState chan<- elevator.Elevator) {
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
			ch_elevatorState <- *e
			time.Sleep(time.Millisecond * 50)
			break
		} else {
			e.Requests[btnFloor][int(btnType)] = true
			request.RequestChooseDirection(e)
			elevio.SetMotorDirection(e.Dir)
			e.Behave = elevator.Moving
			ch_elevatorState <- *e
			time.Sleep(time.Millisecond * 50)
			break
		}

	}
}
