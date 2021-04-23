package watchdog

import (
	"time"
)

func Watchdog(seconds int, ch_pet chan bool, ch_bark chan bool) {
	watchdogTimer := time.NewTimer(time.Duration(seconds) * time.Second)
	for {
		select {
		case <-ch_pet:
			watchdogTimer.Reset(time.Duration(seconds) * time.Second)
		case <-watchdogTimer.C:
			ch_bark <- true
			watchdogTimer.Reset(time.Duration(seconds) * time.Second)
		}
	}
}
