package engine

import (
	"fmt"
	"github.com/teo-mateo/bittrader/api"
	"github.com/teo-mateo/bittrader/data"
	"github.com/teo-mateo/bittrader/data/models"
	"log"
	"strconv"
	"time"
	"github.com/teo-mateo/bittrader/broker"
	"github.com/teo-mateo/bittrader/timesource"
	"github.com/montanaflynn/stats"
)

const (
	Idle   = iota
	ToSell //engine decided it's time to sell (price above position high)
	Selling
	ToBuy
	Buying
)

type PriceRange struct {
	Top    float64
	Bottom float64
}

//// Checks if the current rate should trigger a position action
//func (position *Position) Check(t time.Time, price float64) {
//
//	//fmt.Printf("Check position %d (%f %s, %f %s) @ %v, %f\n", position.NrCrt, position.Crypto, "ETH", position.Fiat, "EUR", t, price)
//	if position.FitForSell(price) {
//		//sell if possible
//		//fmt.Printf("Position %d should sell @ %f\n", position.NrCrt, price)
//
//	} else if position.FitForBuy(price) {
//		//buy if possible
//		//fmt.Printf("Position %d should buy @ %f\n", position.NrCrt, price)
//		if position.Fiat == 0.0 && position.Fiat == 0.0 {
//			//fmt.Printf("Position %d is not initialized. Buying in on empty positions is not supported.\n", position.NrCrt)
//			return
//		}
//
//		if position.Crypto > 0.0 {
//			return
//		}
//
//	}
//}

type TradingEngine struct {
	startingPrice float64
	Investment    float64
	Sim models.Price1mSim
	Range         PriceRange
	BankCrypto    float64
	InitialCrypto float64
	BankFiat      float64
	MaxPositions  int
	Positions     []*Position
	Runtime 	  RuntimeData
	CommissionSpent float64

}

type RuntimeData struct{
	LastPrice 	  float64
	LastTime	  time.Time
	TradingDays	  int
}

func (engine *TradingEngine) DistributeBuyIn(time time.Time, price float64, up bool, down bool) {

	if MONTHLY_BUYIN > 0.0 {

		var position *Position = nil
		for _, pos := range engine.Positions {
			if pos.Low <= price && pos.High > price{
				position = pos
				break
			}
		}

		if position != nil{
			next := position.Next
			prev := position.Previous

			f := MONTHLY_BUYIN * 0.1
			MONTHLY_BUYIN -= f
			position.Fiat += f

			for {

				if MONTHLY_BUYIN <=1.0 || (next == nil && prev == nil){

					newCrypto := MONTHLY_BUYIN / price
					position.Crypto += newCrypto
					engine.InitialCrypto += newCrypto
					break
				}

				distributed := false

				if down {
					if prev != nil {
						f := MONTHLY_BUYIN * 0.1
						MONTHLY_BUYIN -= f

						newCrypto := f/price
						prev.Crypto += newCrypto
						prev = prev.Previous

						engine.InitialCrypto += newCrypto
						distributed = true
					}
				}

				if up {
					if next != nil {
						f := MONTHLY_BUYIN * 0.1
						MONTHLY_BUYIN -= f

						newCrypto := f/price
						next.Crypto += newCrypto
						next = next.Next

						engine.InitialCrypto += newCrypto
						distributed = true
					}
				}

				if !distributed{
					break
				}
			}
		}

		event := NewRedistributeEvent(time, price, &Redistribution{})
		broker.DispatchAny(event)
	}
}


// RedistributePositions runs periodically and redistributes the positions' fiat/crypto to the innermost positions
func (engine *TradingEngine) RedistributePositions(time time.Time, price float64) {

	price_low := (price - price*MIDDLE_PC/ 100.0)
	price_high := (price + price *MIDDLE_PC/ 100.0)

	var ok []*Position = make([]*Position, 0, 0)
	var nok []*Position = make([]*Position, 0, 0)

	for _, position := range engine.Positions{

		if position.Low <= price && position.High > price {
			//current position
			ok = append(ok, position)
		} else if position.High < price {
			//position is below price
			if position.High >= price_low {
				ok = append(ok, position)
			} else if position.Crypto > 0.0 || position.Fiat > 0.0 {
				nok = append(nok, position)
			}
		} else if position.Low > price {
			//position is above price
			if position.Low <= price_high {
				ok = append(ok, position)
			} else if position.Crypto > 0.0 || position.Fiat > 0.0 {
				nok = append(nok, position)
			}
		}
	}

	distributeCrypto := 0.0
	distributeFiat := 0.0

	redist := Redistribution{
		CryptoFrom:make(map[int]float64),
		CryptoTo:make(map[int]float64),
		FiatFrom:make(map[int]float64),
		FiatTo:make(map[int]float64),
	}

	for _, position := range nok {
		pluckCrypto := position.Crypto * PLUCK_PC / 100.0
		if pluckCrypto * price <= 0.5{
			//pluck all
			pluckCrypto = position.Crypto
		}

		distributeCrypto += pluckCrypto
		position.Crypto -= pluckCrypto
		if pluckCrypto > 0.0 {
			redist.CryptoFrom[position.NrCrt] = pluckCrypto
		}


		pluckFiat := position.Fiat * PLUCK_PC / 100.0
		if pluckFiat <= 0.05 {
			//pluck all
			pluckFiat = position.Fiat
		}
		distributeFiat += pluckFiat
		position.Fiat -= pluckFiat
		if pluckFiat > 0.0 {
			redist.FiatFrom[position.NrCrt] = pluckFiat
		}
	}

	for _, position := range ok {

		dfiat := distributeFiat / float64(len(ok))
		dcrypto := distributeCrypto / float64(len(ok))

		position.Fiat += dfiat
		redist.FiatTo[position.NrCrt] = dfiat

		position.Crypto += dcrypto
		redist.CryptoTo[position.NrCrt] = dcrypto
	}

	event := NewRedistributeEvent(time, price, &redist)
	broker.DispatchAny(event)

}

func (engine *TradingEngine) FilterBuyPositions(price float64) []*Position {

	result := make([]*Position, 0)
	for _, position := range engine.Positions {
		if position.Fiat > 0.0 && price <= position.Low {
			result = append(result, position)
		}
	}
	return result
}

func filterOnlyWithCrypto(result []*Position) []*Position {
	result2 := make([]*Position, 0)
	for _, position := range result {
		if position.Crypto > 0.0 {
			result2 = append(result2, position)
		}
	}
	return result2
}

func (engine *TradingEngine) FilterSellPositions(t time.Time, price float64) ([]*Position, bool) {
	result := make([]*Position, 0)
	for _, position := range engine.Positions {

		if price >= position.High  + 0.004 * position.High {
			result = append(result, position)
		}
	}

	if len(result) == 0 {
		return result, false
	}

	//profits := engine.BankCrypto*price - engine.Investment
	//profitsCrypto := profits / price
	//profitsCryptoAssigned := profitsCrypto * 0.1

	assignedProfit := false

	//if profits > 0.000001 {
	//
	//	////assign the profit to the upmost position
	//	//upmostPosition := result[len(result)-1]
	//	//upmostPosition.Crypto += profitsCrypto
	//	//engine.BankCrypto = engine.BankCrypto - profitsCrypto
	//	//assignedProfit = true
	//	//
	//	//event := NewAssignEvent(t, price, upmostPosition)
	//	//broker.DispatchAny(event)
	//	//
	//	//fmt.Printf("%v At %.3f -> PROFIT %f (%.5f EUR) --> Assign %f to Position %d [%.3f %.3f], Bank crypto: %f\n", t, price, profitsCrypto, profits, profitsCryptoAssigned, upmostPosition.NrCrt, upmostPosition.Low, upmostPosition.High, engine.BankCrypto)
	//	//time.Sleep(OP_SLEEP)
	//}

	return filterOnlyWithCrypto(result), assignedProfit
}

func (engine *TradingEngine) GenerateSimulation() models.Simulation {
	return models.Simulation{
		Tradepair:        engine.Sim.Tradepair,
		Investment:       engine.Investment,
		RangeBottom:      engine.Range.Bottom,
		RangeTop:         engine.Range.Top,
		MaxPositions:     float64(engine.MaxPositions),
		TrendRate:        MOVE_PROFIT_ABOVE_PC,
		PositionStepover: POSITION_STEPOVER,
		Positions:        make([]*models.Position, engine.MaxPositions),
	}
}

//func (engine *TradingEngine) Step() float64 {
//	return (engine.Range.Top - engine.Range.Bottom) / float64(engine.MaxPositions)
//}

func (engine *TradingEngine) positionsCrypto() float64 {
	c := 0.0
	for _, p := range engine.Positions {
		c = c + p.Crypto
	}
	return c
}

func (engine *TradingEngine) totalCrypto() float64 {
	return engine.BankCrypto + engine.positionsCrypto()
}

func (engine *TradingEngine) positionsFiat() float64 {
	f := 0.0
	for _, p := range engine.Positions {
		f = f + p.Fiat
	}
	return f
}

func (engine *TradingEngine) totalFiat() float64 {
	return engine.BankFiat + engine.positionsFiat()
}

func (engine *TradingEngine) Prepare() {

	engine.CommissionSpent = 0.0
	engine.Runtime = RuntimeData{}

	timechannel := timesource.GetTimeChannel()
	timesource.Play()
	for {
		t, ok := <- timechannel
		if ok {
			engine.handleTime(t)
		} else {
			fmt.Println("Done")
			break
		}
	}

	engine.printSummary()
}

var profitInfo = make(map[time.Time]map[*Position]float64)
var todaysPrices = make([]float64, 0)

func (engine *TradingEngine) handleTime(t time.Time){
	price := myapi.GetPriceAtTime(engine.Sim.Id, t)

	engine.Runtime.LastTime = t
	engine.Runtime.LastPrice = price
	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)


	if _, ok := profitInfo[today]; !ok {

		//summary for yesterday
		yesterday := today.Add(-1 * 24 * time.Hour)
		total := 0.0
		if _, ok = profitInfo[yesterday]; ok {

			//new day: daily report for yesterday
			s := fmt.Sprintf("----------> DAYLY REPORT for %v\n", yesterday)
			fmt.Printf(s)
			for pos, profit := range profitInfo[yesterday]{
				total += profit
				fmt.Printf("----------> Position %d got profit of %.3f EUR \n", pos.NrCrt, profit)
			}
			if total > 0{
				fmt.Printf("----------> TOTAL: %.3f EUR \n", total)
				//time.Sleep(2 * time.Second)
			}

			//new day: compute standard deviation of yesterday's prices
			if len(todaysPrices) > 0{
				//compute std deviation
				sdev, err := stats.StandardDeviation(todaysPrices)
				if err != nil{
					log.Panic(err)
				}
				fmt.Printf("Yesterday's standard deviation was %.3f\n\n", sdev)
			}
			todaysPrices = make([]float64, 0)
		}

		profitInfo[today] = make(map[*Position]float64)
	}

	if price > 0.0 {

		todaysPrices = append(todaysPrices, price)

		bought := false
		positionsToBuy := engine.FilterBuyPositions(price)
		for _, p := range positionsToBuy {
			b, profit := p.Buy(t, price)
			if b {
				profitInfo[today][p] += (profit * price)
			}
			bought = b && bought
		}

		sold := false
		positionsToSell, assignedProfit := engine.FilterSellPositions(t, price)
		for _, p := range positionsToSell {
			sold = p.Sell(t, price) && sold
		}

		if !bought && !sold && !assignedProfit && t.Minute() % 50 == 0 {
			event := NewPriceEvent(t, price, engine)
			broker.DispatchAny(event)
		}
	}

	if t.Weekday()== time.Monday && t.Hour() == 0 && t.Minute() == 0 {
		//engine.RedistributePositions(t, price)
	}

	if t.Day() == 1 && t.Hour() == 0 && t.Minute() == 0{
		bck := MONTHLY_BUYIN
		engine.DistributeBuyIn(t, price, true, false)
		MONTHLY_BUYIN = bck
		engine.Investment = engine.Investment + MONTHLY_BUYIN
	}
}



func InitTradingEngine(start time.Time, end time.Time, investment float64, sim_id int, range_bottom float64, range_top float64, maxpositions int) *TradingEngine {

	//set up time source
	timesource.Init(timesource.TimeSourceInfo{
		StartTime: start,
		EndTime: end,
		Step:TIME_INCREMENT,
		OpSleep:OP_SLEEP,
	})

	sim, err := data.GetSim(sim_id)
	if err != nil{
		log.Panic(err)
	}

	startingPrice := myapi.GetPriceAtTime(sim_id, start)
	bankCrypto := investment / startingPrice

	e := TradingEngine{}
	e.startingPrice = startingPrice
	e.Investment = investment
	e.Sim = sim
	e.Range = PriceRange{Bottom: range_bottom, Top: range_top}
	e.BankCrypto = bankCrypto
	e.BankFiat = 0
	e.MaxPositions = maxpositions
	e.InitialCrypto = bankCrypto

	//generate positions
	var previousPosition *Position = nil

	e.Positions = make([]*Position, 0, 10)
	for i := 0; i == 0 || previousPosition.High < range_top; i++ {

		newpos := Position{
			NrCrt:  i + 1,
			Engine: &e,
		}

		//link back
		newpos.Previous = previousPosition

		newpos.Low = range_bottom

		//link fwd
		if previousPosition != nil {
			previousPosition.Next = &newpos
			newpos.Low = previousPosition.High
		}

		//compute high
		h1 := (POSITION_INCREMENT_PC*newpos.Low)/100 + newpos.Low
		newpos.High = h1
		newpos.Low = newpos.Low - newpos.Low * POSITION_STEPOVER / 100

		//one last thing
		previousPosition = &newpos
		e.Positions = append(e.Positions, &newpos)
	}

	for _, position := range e.Positions {
		prevNrCrt := ""
		if position.Previous != nil {
			prevNrCrt = strconv.Itoa(position.Previous.NrCrt)
		}
		nextNrCrt := ""
		if position.Next != nil {
			nextNrCrt = strconv.Itoa(position.Next.NrCrt)
		}

		position.Crypto = e.BankCrypto / float64(len(e.Positions))

		fmt.Printf("( %s <- %d -> %s ) [%.6f %.6f]   %.3f %s   %.3f %s\n", prevNrCrt, position.NrCrt, nextNrCrt, position.Low, position.High, position.Crypto, sim.Tradepair.Cfrom, position.Fiat, sim.Tradepair.Cto)
	}
	e.BankCrypto = 0.0

	return &e
}
