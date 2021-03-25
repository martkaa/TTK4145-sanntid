package timer

import (
	"Project/elevator"
	"fmt"
	"time"
)

func TimerDoor(sec int, timerChan chan<- bool, e *elevator.Elevator) {
	e.TimerCount += 1
	fmt.Println(e.TimerCount)
	time.Sleep(time.Duration(sec) * time.Second)
	timerChan <- true
}
