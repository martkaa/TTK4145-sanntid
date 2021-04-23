package fsm

import (
	"Project/config"
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/request"
	"time"
)

// Final state machine to run the local elevator.
func Fsm(
	ch_orderChan chan elevio.ButtonEvent,
	ch_elevatorState chan<- elevator.Elevator,
	ch_clearLocalHallOrders chan bool,
	ch_arrivedAtFloors chan int,
	ch_obstruction chan bool,
	ch_timerDoor chan bool) {

	elev := elevator.InitElev()
	e := &elev

	elevio.SetDoorOpenLamp(false)
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

	// Initialize timers
	doorTimer := time.NewTimer(time.Duration(config.DoorOpenDuration) * time.Second)
	timerUpdateState := time.NewTimer(time.Duration(config.StateUpdatePeriodMs) * time.Microsecond)

	for {
		elevator.LightsElev(*e)
		select {
		case order := <-ch_orderChan:
			switch {
			case e.Behave == elevator.DoorOpen:
				if e.Floor == order.Floor {
					doorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
				} else {
					e.Requests[order.Floor][int(order.Button)] = true
				}
			case e.Behave == elevator.Moving:
				e.Requests[order.Floor][int(order.Button)] = true
			case e.Behave == elevator.Idle:
				if e.Floor == order.Floor {
					elevator.LightsElev(*e)
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
					e.Behave = elevator.DoorOpen
					ch_elevatorState <- *e
					break
				} else {
					e.Requests[order.Floor][int(order.Button)] = true
					request.RequestChooseDirection(e)
					elevio.SetMotorDirection(e.Dir)
					e.Behave = elevator.Moving
					ch_elevatorState <- *e
					break
				}
			}
		case floor := <-ch_arrivedAtFloors:
			e.Floor = floor
			switch {
			case e.Behave == elevator.Moving:
				if request.RequestShouldStop(e) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevator.LightsElev(*e)
					request.RequestClearAtCurrentFloor(e)
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
					e.Behave = elevator.DoorOpen
					ch_elevatorState <- *e
				}
			default:
				break
			}
		case <-doorTimer.C:
			switch {
			case e.Behave == elevator.DoorOpen:
				request.RequestChooseDirection(e)
				elevio.SetMotorDirection(e.Dir)
				elevio.SetDoorOpenLamp(false)

				if e.Dir == elevio.MD_Stop {
					e.Behave = elevator.Idle
					ch_elevatorState <- *e
				} else {
					e.Behave = elevator.Moving
					ch_elevatorState <- *e
				}
			default:
				break
			}
		case <-ch_clearLocalHallOrders:
			request.RequestClearHall(e)
		case obstruction := <-ch_obstruction:
			if e.Behave == elevator.DoorOpen && obstruction {
				doorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
			}
		case <-timerUpdateState.C:
			ch_elevatorState <- *e
			timerUpdateState.Reset(time.Duration(config.StateUpdatePeriodMs) * time.Millisecond)
		}
	}
}
