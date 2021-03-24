package main

import (
	"fmt"

	"Project/elevator"
	"Project/elevio"
	"Project/fsm"
)

func main() {

	elev := elevator.InitElev(elevator.NumFloors, elevator.NumButtons)

	e := &elev

	elevio.Init("localhost:12345", elevator.NumFloors)

	//elevio.SetMotorDirection(e.Dir)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	timerChan := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	for {
		fmt.Println(e.Behave)
		elevator.LightsElev(*e)
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button, e, timerChan)

		case f := <-drv_floors:
			e.Floor = f
			fsm.FsmOnFloorArrival(e, timerChan)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(e.Dir)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < elevator.NumFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		case <-timerChan:
			fsm.FsmOnDoorTimeout(e)
		}
	}
}
