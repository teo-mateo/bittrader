package engine

import (
	"fmt"
	"github.com/teo-mateo/bittrader/timesource"
)

func (engine *TradingEngine) printSummary() {

	lastprice := engine.Runtime.LastPrice

	positionsCrypto := 0.0
	for i, position := range engine.Positions {
		positionsCrypto = positionsCrypto + position.Crypto
		pc_gain := (position.High - position.Low) * 100 / position.High
		fmt.Printf("%d [%.3f %.3f] **%.3f** %f %s (%f g) %.2f %s %d PP: %.5f\n", i+1, position.Low, position.High, pc_gain, position.Crypto, engine.Sim.Tradepair.Cfrom, 0.0, position.Fiat, engine.Sim.Tradepair.Cto, position.HitCount, position.TotalProfit)
	}

	fmt.Println(" ")
	totalHits := 0
	for _, position := range engine.Positions {
		totalHits = totalHits + position.HitCount
	}
	fmt.Printf("Total hits: %d\n", totalHits)
	fmt.Printf("Bank C1: %f %s\n", engine.BankCrypto, engine.Sim.Tradepair.Cfrom)
	fmt.Printf("Bank C2: %f %s\n", engine.BankFiat, engine.Sim.Tradepair.Cto)
	fmt.Printf("%s in positions: %f (%.3f %s)\n", engine.Sim.Tradepair.Cfrom, positionsCrypto, positionsCrypto*lastprice, engine.Sim.Tradepair.Cto)
	fmt.Printf("%s in positions: %.3f\n", engine.Sim.Tradepair.Cto, engine.positionsFiat())

	allCrypto := positionsCrypto + engine.BankCrypto

	fmt.Printf("%s Total: %f\n", engine.Sim.Tradepair.Cfrom, allCrypto)
	fmt.Println("**********************")
	fmt.Printf("%.3f --> %.3f\n", engine.startingPrice, lastprice)

	assetsAfterTrading := engine.BankFiat + allCrypto * lastprice + engine.positionsFiat()
	assetsNoTrading := lastprice * engine.InitialCrypto
	profitPercent := (assetsAfterTrading - assetsNoTrading) * 100.0 / assetsNoTrading

	totalTradingDays := timesource.RunningDays()

	s := fmt.Sprintf("%.3f (at %.3f) --> %.3f (at %.3f) (vs %.3f) (%.3f%%) over %.2f days\n", engine.Investment, engine.startingPrice, assetsAfterTrading, lastprice, assetsNoTrading, profitPercent, totalTradingDays)
	fmt.Println(s)

	fmt.Printf("Spent on comission: %.6f %s\n", engine.CommissionSpent, engine.Sim.Tradepair.Cfrom)
}