package orchestrator

import (
	"time"
	engine2 "github.com/teo-mateo/bittrader/engine"
	"github.com/teo-mateo/bittrader/timesource"
)

var OP_SLEEP = 10 * time.Millisecond
var RunParams *EngineParams

var e *engine2.TradingEngine

func init(){
}

var onpause = false

func Pause(){
	onpause = true
	timesource.Pause()
}

func Play(){
	if !onpause {
		prepare()
	}
	timesource.Play()
}

func Stop(){
	timesource.Stop()
}

