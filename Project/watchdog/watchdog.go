package watchdog

import (
	"Project/network/peers"
	"time"
)

func WatchdogLostConnection(seconds int, ch_peerUpdate chan peers.PeerUpdate, ch_resetWatchdog chan string, ch_watchdogLostConnection chan string) {
	for {
		peer := <-ch_peerUpdate
		if peer.New != "" {
			go watchdog(peer.New, seconds, ch_resetWatchdog, ch_watchdogLostConnection)
		}
		for _, ID := range peer.Peers {
			for range peer.Peers {
				ch_resetWatchdog <- ID
			}
		}
	}
}

func WatchdogElevatorStuck(seconds int, ch_elevStuck chan bool, ch_watchdogElevatorStuck chan bool) {
	counter := 0
	for {
		time.Sleep(time.Second)
		select {
		case elevStuck := <-ch_elevStuck:
			if !elevStuck {
				counter = 0
			}
		default:
			counter += 1
			if counter == seconds {
				ch_watchdogElevatorStuck <- true
				return
			}
		}
	}
}

func watchdog(ID string, seconds int, ch_resetWatchdog chan string, ch_watchdogAlarm chan string) {
	counter := 0
	for {
		time.Sleep(time.Second)
		select {
		case watchdogID := <-ch_resetWatchdog:
			if watchdogID == ID {
				counter = 0
			}
		default:
			counter += 1
			if counter == seconds {
				ch_watchdogAlarm <- ID
				return
			}
		}
	}
}

/*secondsseconds


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
