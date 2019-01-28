package main

import (
	"fmt"
	"github.com/teo-mateo/bittrader/data"
	"log"
	"time"
	"github.com/teo-mateo/bittrader/consts"
	"github.com/teo-mateo/bittrader/engine"
	"github.com/teo-mateo/bittrader/importz"
	//"github.com/teo-mateo/bittrader/api"
	"github.com/teo-mateo/bittrader/orchestrator"
)

var SIM_ID = 1

func importethbtc(){
	importz.ImportData(171, 30, "ticks_ethbtc_onemin.json")
}

func main() {
	fmt.Println("BitTrader started")

	startHTTPServer()

	//myapi.GetAssetPairs()
	//return

	//time.Sleep(2 * time.Second)

	//updateDB()

	//runBatch()

	runSingle()

	//importethbtc()

	//generateDummyPrice1m()

	//generatePrice1h()



	for {
		fmt.Printf(".")
		time.Sleep(5000*time.Millisecond)
	}


}

func updateDB() {

	startHTTPServer()

	SIM_ID = consts.SIM_XTHXXRP_ID_ORIGINAL
	getTradesTakesTooFuckingLong()
	generatePrice1m()

	//SIM_ID = consts.SIM_ETHEUR_ID_ORIGINAL
	//getTradesTakesTooFuckingLong()
	//generatePrice1m()
	//
	//SIM_ID = consts.SIM_LTCEUR_ID_ORIGINAL
	//getTradesTakesTooFuckingLong()
	//generatePrice1m()
	//
	//SIM_ID = consts.SIM_BTCEUR_ID_ORIGINAL
	//getTradesTakesTooFuckingLong()
	//generatePrice1m()
	//
	//SIM_ID = consts.SIM_XRPEUR_ID_ORIGINAL
	//getTradesTakesTooFuckingLong()
	//generatePrice1m()

}

func runSingle() {

	//// __ETHBTC__
	//INVESTMENT := 1.0
	//RANGE_BOTTOM := 0.03
	//RANGE_TOP := 0.1
	//MAXPOSITIONS := 50

	// __LTC__
	//INVESTMENT := 1000.0
	//RANGE_BOTTOM := 36.0
	//RANGE_TOP := 100.0
	//MAXPOSITIONS := 500

	// __ETH__
	INVESTMENT := 1000.0
	RANGE_BOTTOM := 220.0
	RANGE_TOP := 400.0
	MAXPOSITIONS := 100

	// __BTC__
	//INVESTMENT := 1000.0
	//RANGE_BOTTOM := 100.0
	//RANGE_TOP := 6000.0
	//MAXPOSITIONS := 1100

	// __XRP__
	//INVESTMENT := 1000.0
	//RANGE_BOTTOM := 0.1
	//RANGE_TOP := 0.35
	//MAXPOSITIONS := 75

	//engine.POSITION_STEPOVER = 0
	//engine.MOVE_PROFIT_ABOVE_PC = 0.1
	//engine.SAVE_TRADES = false
	//engine.USE_ONLY_PROFITS = true

	start := time.Date(2017, 8, 18, 0, 1, 0, 0, time.Local)
	end := time.Date(2017, 11, 10, 23, 0, 0, 0, time.Local)

	//start := time.Date(2030, 1, 1, 0, 0, 0, 0, time.Local)
	//end := time.Date(2030, 1, 1, 1, 5, 0, 0, time.Local)
	//INVESTMENT := 1000.0
	//RANGE_BOTTOM := 80.0
	//RANGE_TOP := 120.0
	//MAXPOSITIONS := 100

	orchestrator.PrepareSimulation(orchestrator.EngineParams{
		SimInfo:orchestrator.SimulationInfo{
			StartTime:start,
			EndTime:end,
			MaxPositions:MAXPOSITIONS,
			RangeTop:RANGE_TOP,
			RangeBottom:RANGE_BOTTOM,
			Investment:INVESTMENT,
			SimId:SIM_ID,
		},
		OpSleep:engine.OP_SLEEP,
		TimeIncrementMs:engine.TIME_INCREMENT,
		Redistribution: orchestrator.RedistributeConfig{
			MiddlePc:engine.MIDDLE_PC,
			PluckPc:engine.PLUCK_PC,
		},
		TrendRate:0.0,
		MinProfitFiat:engine.MIN_PROFIT_FIAT,
		MontlyBuyin:engine.MONTHLY_BUYIN,
		PositionIncrementPc:engine.POSITION_INCREMENT_PC,
		})

	orchestrator.Play()
}

func generateDummyPrice1m() {
	data.GenerateDummyPrice1m(SIM_ID, 260, 30)
}

func generatePrice1m() {
	data.GeneratePrice1m(SIM_ID)
}

func generatePrice1h() {
	data.GeneratePrice1h(SIM_ID)
}

func getTradesTakesTooFuckingLong() {
	cutoff := time.Now()
	for {
		lasttradetime, err := data.UpdateTrades(SIM_ID)
		if err != nil {
			fmt.Println(err)
			time.Sleep(10 * time.Second)
			continue
		}
		fmt.Println(lasttradetime)
		if lasttradetime.After(cutoff) {
			break
		}
	}
}

func getTradePairs() {
	pairs, err := data.GetTradePairs()
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range pairs {
		fmt.Println(p.Nicename)
	}
}
