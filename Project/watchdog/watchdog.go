package watchdog

import (
	"time"
)

/*
import (
	"time"
)

elevators := make([]*config.DistributorElevator, 0)

//Fuc to check if there are any hall orders
func hasOrders(elevState elevators) bool {
	for f := range elevState.Requests {
		for b := range elevState.Requests[f] {
			if elevState.Requests[f][b] {
				return true
			}
		}
	}
	return false
}

//Watchdog - monitor that elevators are moving, if not assign to local elevator
func watchdog(timeOutC chan<- bool, elevState chan<- elevators, timeout time.Duration) {
	watchdogEnabled := false
	watchdogTimer := time.NewTimer(timeout)

	for {
		select {
		case newState := <-elevState:
			//Enables as long as there exists hall orders
			watchdogEnabled = hasOrders(newState)

			//Reset timer if any elev has moved
			for newElevID, newElevFloor := range newState.elevators {
				if floor, ok := floorMap[newElevID]; ok {
					if floor != newElevFloor {
						if watchdogTimer.Stop() {
							watchdogTimer.Reset(timeout)
						}
					}
				}
				floorMap[newElevID] = newElev.Floor
			}
		//Watchdog timed out, alert distributor
		case <-watchdogTimer.C:
			timeout <- true
			watchdogTimer.Reset(timeout)
		default:
			if !watchdogEnabled && watchdogTimer.Stop() {
				watchdogTimer.Reset(timeout)
			}
		}
	}
}

*/

func WatchdogElevatorStuck(seconds int, ch_elevStuck chan bool, ch_watchdogElevatorStuck chan bool) {
	watchdogTimer := time.NewTimer(time.Duration(seconds) * time.Second)
	for {
		select {
		case elevStuck := <-ch_elevStuck:
			if !elevStuck {
				watchdogTimer.Reset(time.Duration(seconds) * time.Second)
			}
		case <-watchdogTimer.C:
			ch_watchdogElevatorStuck <- true
			watchdogTimer.Reset(time.Duration(seconds) * time.Second)
		}
	}
}
