package timer

import (
	"time"

	"Project/elevator"
)

func TimerDoor(sec int, ch_timerDoor chan<- bool, e *elevator.Elevator) {
	e.TimerCount += 1
	time.Sleep(time.Duration(sec) * time.Second)
	ch_timerDoor <- true
}

func TimerUpdateState(millisec int, ch_timerUpdateState chan bool) {
	for {
		time.Sleep(time.Duration(millisec) * time.Millisecond)
		ch_timerUpdateState <- true
	}
}
