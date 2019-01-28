package engine

import (
	"fmt"
	"github.com/teo-mateo/bittrader/data"
	"log"
	"time"
	"github.com/teo-mateo/bittrader/broker"
)


type Position struct {
	NrCrt       int
	Id          int64
	Low         float64
	High        float64
	Crypto      float64
	Fiat        float64
	Engine      *TradingEngine
	Status      int
	OrderId     string
	HitCount    int
	TotalProfit float64
	//NextProfit  float64
	SoldCrypto float64
	Previous   *Position
	Next       *Position
}

//func (position *Position) FitForBuy(price float64) bool {
//
//	//position is fit for buy if
//	//2. price below position.low
//
//	result := false
//
//	if position.Previous == nil && price <= position.Low {
//		result = true
//	} else if position.Previous != nil && price > position.Previous.Low {
//		result = true
//	}
//
//	fmt.Printf("PRICE: %.3f Position %d [%.3f %.3f] BUY? %v\n", price, position.NrCrt, position.Low, position.High, result)
//	return result
//
//}
//
//func (position *Position) FitForSell(price float64) bool {
//	//position is fit for sell if
//	//2. price above position.high
//
//	result := false
//
//	if position.High >= price && ((position.Next == nil) || (position.Next != nil && position.Next.High > price)) {
//		result = true
//	}
//
//	fmt.Printf("PRICE: %.3f Position %d [%.3f %.3f] SELL? %v\n", price, position.NrCrt, position.Low, position.High, result)
//	return result
//}

func (position *Position) AsPositionInfo() PositionInfo {

	return PositionInfo{
		NrCrt:    position.NrCrt,
		HitCount: position.HitCount,
		Low:      position.Low,
		High:     position.High,
		Crypto:   position.Crypto,
		Fiat:     position.Fiat,
	}
}

func (position *Position) Buy(t time.Time, price float64) (resbool bool, profit float64) {
	if position.Fiat == 0.0 {
		return false, 0.0
	} else {
		//buy
		position.Crypto += position.Fiat / price
		profit := position.Crypto - position.SoldCrypto
		position.HitCount = position.HitCount + 1
		position.TotalProfit = position.TotalProfit + profit

		profitFiat := profit * price

		fmt.Printf("Position %d: %s %.3f bought %f %s @ %f %s, will sell @ %f. PROFIT: %.6f\n", position.NrCrt, t.Format("2006-01-02T15:04:05"), price, position.Crypto, position.Engine.Sim.Tradepair.Cfrom, position.Fiat, position.Engine.Sim.Tradepair.Cto, position.High, profitFiat)
		position.Fiat = 0.0

		//update db after BUY; has crypto
		if SAVE_TRADES {
			err := data.PositionBuy(position.Id, t, price, position.Crypto, profit)
			if err != nil {
				log.Panic(err)
			}
		}

		//distribute crypto
		cryptoProfit := position.Crypto - position.SoldCrypto
		if MOVE_PROFIT_ABOVE_PC > 0 && position.Next != nil && position.Next.Next != nil && position.Next.Next.Next != nil{
			toMove := cryptoProfit * MOVE_PROFIT_ABOVE_PC / 100
			position.Next.Next.Next.Crypto += toMove
			position.Crypto -= toMove

			fmt.Printf("Moved above: %.3f to %d\n", toMove, position.Next.Next.NrCrt)
			event := NewAssignEvent(t, price, position.Next.Next)
			broker.DispatchAny(event)
		}

		time.Sleep(OP_SLEEP)

		event := NewBuyEvent(t, price, position, position.Crypto, position.Engine.Sim.Tradepair.Cfrom, 0.0)
		broker.DispatchAny(event)
		return true, profit
	}
}

func (position *Position) Sell(t time.Time, price float64) (resbool bool) {
	commission := 0.0
	if position.Crypto == 0.0 && position.Fiat == 0.0 {
		crypto := 0.0

		if USE_ONLY_PROFITS {
			//only use profits
			//crypto = MOVE_PROFIT_ABOVE_PC * (position.Engine.BankCrypto*price - position.Engine.Investment) / price
		} else {
			//use all amount
			crypto = position.Engine.BankCrypto / float64(position.Engine.MaxPositions)
		}

		if crypto <= 0.0 {
			//fmt.Printf("Not enough funds for position %d\n", position.NrCrt)
			return
		}
		position.Crypto += crypto
		position.Engine.BankCrypto = position.Engine.BankCrypto - position.Crypto
	}

	//crypto to sell
	crypto := position.Crypto

	if position.Fiat > 0.0 {
		//return
	}

	commission = crypto * COMMISSION_PC
	position.Engine.CommissionSpent += commission
	fiat := (crypto - commission) * price
	nextProfit := fiat/position.Low - crypto

	if nextProfit*position.Low < MIN_PROFIT_FIAT {
		fmt.Printf("Position %d [%.3f %.3f] %f: %s %.3f Not enough profit: %f \n", position.NrCrt, position.Low, position.High, position.Crypto, t, price, nextProfit*position.Low)
		return
	}

	fmt.Printf("Position %d: %s %.3f sold %f %s for %f %s\n", position.NrCrt, t.Format("2006-01-02T15:04:05"), price, position.Crypto, position.Engine.Sim.Tradepair.Cfrom, fiat, position.Engine.Sim.Tradepair.Cto)

	//actual sale
	position.Fiat += fiat

	//reset crypto
	position.Crypto = 0.0

	//how much we sold
	position.SoldCrypto = crypto - commission

	//update db after SELL: has fiat.
	if SAVE_TRADES {
		err := data.PositionSell(position.Id, t, price, position.Fiat)
		if err != nil {
			log.Panic(err)
		}
	}

	time.Sleep(OP_SLEEP)

	event := NewSellEvent(t, price, position, position.Crypto, position.Engine.Sim.Tradepair.Cfrom)
	broker.DispatchAny(event)

	return true

}
