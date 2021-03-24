package timer

import (
	"time"
)

func TimerDoor(sec int, timerChan chan<- bool) {
	time.Sleep(time.Duration(sec) * time.Second)
	timerChan <- true
}
