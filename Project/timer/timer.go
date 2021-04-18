package timer

import (
	"../elevator"
	"time"
)

func TimerDoor(sec int, ch_timerChan chan<- bool, e *elevator.Elevator) {
	e.TimerCount += 1
	time.Sleep(time.Duration(sec) * time.Second)
	ch_timerChan <- true
}
