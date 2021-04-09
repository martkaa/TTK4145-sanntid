package distributor

import (
	"Project/elevator"
	"Project/elevio"
)

type DistributorElevator struct {
	Id       int
	Floor    int
	Dir      elevio.MotorDirection
	Requests [][]bool
	Behave   elevator.Behaviour
}

type Request struct {
	Floor int
	btn   elevio.ButtonType
	// idé å inkludere dette:?
	// DesignatedElev	int
	// Done				bool
}

func Distributor(elevatorChannel chan<- chan Request) {

	elevators := make([]DistributorElevator, 0)

	elevatorUpdate := make(chan DistributorElevator)
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
		case floor, btn := <-newInternalRequest:
			//Start go routine that calculates which elevator to execute the request
		case elevatorChannel <- assignedRequest:
		}

	}
}
