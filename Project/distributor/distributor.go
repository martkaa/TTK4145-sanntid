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
