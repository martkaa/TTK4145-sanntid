package distributor

import (
	"Project/cost"
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
	Floor            int
	Btn              elevio.ButtonType
	DesignatedElevID int
	Done             bool
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
		case r := <-newInternalRequest:
			go cost.CostCalculator()
		case elevatorChannel <- assignedRequest:
		}

	}
}
