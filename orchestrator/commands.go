package orchestrator

import (
	"time"
	"github.com/teo-mateo/bittrader/engine"
)

type Command struct {
	Instruction string
	Params *EngineParams
}

type EngineParams struct {
	PositionIncrementPc float64
	TrendRate float64
	TimeIncrementMs time.Duration
	OpSleep time.Duration
	MinProfitFiat float64
	MontlyBuyin float64
	Redistribution RedistributeConfig
	SimInfo SimulationInfo
}

type RedistributeConfig struct {
	MiddlePc float64
	PluckPc float64
}

type SimulationInfo struct {
	SimId int
	StartTime time.Time
	EndTime time.Time
	Investment float64
	RangeBottom float64
	RangeTop float64
	MaxPositions int
}

var prepare func()

func PrepareSimulation(params EngineParams){
	onpause = false
	prepare = func(){
		e = engine.InitTradingEngine(
			params.SimInfo.StartTime,
			params.SimInfo.EndTime,
			params.SimInfo.Investment,
			params.SimInfo.SimId,
			params.SimInfo.RangeBottom,
			params.SimInfo.RangeTop,
			params.SimInfo.MaxPositions)

		engine.OP_SLEEP = params.OpSleep
		engine.TIME_INCREMENT = params.TimeIncrementMs
		engine.MIDDLE_PC = params.Redistribution.MiddlePc
		engine.PLUCK_PC = params.Redistribution.PluckPc
		engine.MIN_PROFIT_FIAT = params.MinProfitFiat
		engine.MONTHLY_BUYIN = params.MontlyBuyin
		engine.POSITION_INCREMENT_PC = params.PositionIncrementPc
		e.Prepare()
	}
}

func GetEngine() *engine.TradingEngine{
	return e
}