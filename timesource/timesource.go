package timesource

import (
	"time"
	"log"
)

var initialized = false
func checkInitialized(){
	if !initialized {
		log.Panic("timesource not initialized")
	}
}

var channel chan time.Time
var startTime time.Time
var endTime time.Time
var currentTime time.Time
var step time.Duration
var opsleep time.Duration
var running = false

type TimeSourceInfo struct {
	StartTime time.Time
	EndTime time.Time
	Step time.Duration
	OpSleep time.Duration
}

func GetTimeChannel() chan time.Time {
	checkInitialized()
	return channel
}

func Init(info TimeSourceInfo) chan time.Time {
	channel = make(chan time.Time, 10)
	startTime = info.StartTime
	currentTime = info.StartTime
	endTime = info.EndTime
	step = info.Step
	opsleep = info.OpSleep
	initialized = true
	return channel
}

func RunningDays() float64 {
	return currentTime.Sub(startTime).Hours() / 24
}

func Play(){
	checkInitialized()
	go func() {
		running = true
		for {
			channel <- currentTime
			time.Sleep(opsleep)
			if !running {
				break
			}
			currentTime = currentTime.Add(step)
			if currentTime.After(endTime){
				close(channel)
				break
			}
		}
	}()
}

func Stop(){
	checkInitialized()
	running = false
	currentTime = startTime
}

func Pause(){
	checkInitialized()
	running = false
}

