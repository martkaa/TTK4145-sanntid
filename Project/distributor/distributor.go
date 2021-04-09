package distributor

import (
	"Project/cost"
	"Project/elevator"
	"Project/elevio"
)

type Request struct {
	Floor            int
	Btn              elevio.ButtonType
	DesignatedElevID int
	Done             bool
}

func Distributor(elevatorChannel chan<- chan Request) {

	elevators := make([]elevator.Elevator, 0)

	elevatorUpdate := make(chan elevator.Elevator)
	newInternalRequest := make(chan Request)
	assignedRequest := make(chan Request)

	//Run go routines that gets states from network and local elevator
	//and

	for {
		select {
		case newElevator := <-elevatorUpdate:
			for _, elevator := range elevators {
				if elevator.Id == newElevator.Id {
					break
				}
				elevators = append(elevators, newElevator)
			}
		case r := <-newInternalRequest:
			go cost.CostCalculator(r, elevators)
		case elevatorChannel <- assignedRequest:
		}

	}
}



func TimeToIdle(elev Elevator) {
    duration := 0;
    select {
    case elev.behaviour == Idle:
        elev.Dir = requests_chooseDirection(elev);
        if elev.Dir == D_Stop {
            return duration
        }
        break
    case elev.behaviour == Moving:
        duration = duration + TRAVEL_TIME/2
        elev.Floor = elev.Floor + elev.Dir
        break
    case elev.behaviour == DoorOpen:
        duration = duration - DOOR_OPEN_TIME/2
    }
    for{
        if requests_shouldStop(elev){
            elev = requests_clearAtCurrentFloor(elev, NULL)
            duration = duration + DOOR_OPEN_TIME
            elev.Dir = requests_chooseDirection(elev)
            if elev.Dir == D_Stop {
                return duration
            }
        }
        elev.Floor = elev.Floor + elev.direction
        duration = duration += TRAVEL_TIME
    }
}