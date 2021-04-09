package cost

import (
	"Project/distributor"
	"Project/elevator"
	"Project/elevio"
	"Project/request"
)

const TRAVEL_TIME = 10

const NumElevators = 4

// Returnerer heis med lavest kost basert på et straffesystem/poeng
// Må deretter sorteres og delegeres

func CostCalculator(request distributor.Request, elevList [NumElevators]elevator.Elevator, onlineList [NumElevators]bool) int {
	if request.Btn == elevio.BT_Cab {
		return id
	}
	minCost := (elevator.NumButtons * elevator.NumFloors) * NumElevators
	bestElevator := id
	for e := 0; e < NumElevators; e++ {
		if !onlineList[e] {
			// Neglect offline elevators
			continue
		}
		cost := request.Floor - elevList[e].Floor

		if cost == 0 && elevList[e].Behave != elevator.Moving {
			bestElevator = e
			return bestElevator
		}
		if cost < 0 {
			cost = -cost
			if elevList[e].Dir == elevio.MD_Up {
				cost += 3
			}

		} else if cost > 0 {
			if elevList[e].Dir == elevio.MD_Down {
				cost += 3
			}
		}
		if cost == 0 && elevList[e].Behave == elevator.Moving {
			cost += 4
		}
		if elevList[e].Behave == elevator.DoorOpen {
			cost++
		}
		if cost < minCost {
			minCost = cost
			bestElevator = e
		}
	}
	return bestElevator
}

/*func distrubute() {
	var (
		elevList       [NumElevators]elevator.Elevator
		onlineList     [NumElevators]bool
		completedOrder Keypress
	)
	completedOrder.DesignatedElevator = id
	elevList[id] = <-elevatorCh
	updateSyncCh <- elevList
}*/

func TimeToServeRequest(e_old elevator.Elevator, r distributor.Request) int {
	e := e_old
	e.Requests[r.Floor][r.Btn] = true

	arrivedAtRequest := false

	duration := 0

	switch e.Behave {
	case elevator.Idle:
		request.RequestChooseDirection(&e)
		if e.Dir == elevio.MD_Stop {
			return duration
		}
	case elevator.Moving:
		duration += TRAVEL_TIME / 2
		e.Floor += int(e.Dir)
		break
	case elevator.DoorOpen:
		duration -= elevator.DoorOpenDuration / 2
	}

	for {
		if request.RequestShouldStop(&e) {
			request.RequestClearAtCurrentFloor(&e)
			if arrivedAtRequest {
				return duration
			}
			duration += elevator.DoorOpenDuration
			request.RequestChooseDirection(&e)
		}
		e.Floor += int(e.Dir)
		duration += TRAVEL_TIME
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