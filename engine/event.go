package engine

import "time"

type PositionInfo struct {
	NrCrt    int     `json:"nrcrt"`
	Low      float64 `json:"low"`
	High     float64 `json:"high"`
	Crypto   float64 `json:"crypto"`
	Fiat     float64 `json:"fiat"`
	HitCount int     `json:"hitcount"`
}

type SellOrder struct {
	Position PositionInfo `json:"position"`
	Kind     string       `json:"kind"`
	Volume   float64      `json:"volume"`
	Currency string       `json:"currency"`
}

type BuyOrder struct {
	Position PositionInfo `json:"position"`
	Kind     string       `json:"kind"`
	Volume   float64      `json:"volume"`
	Currency string       `json:"currency"`
	Profit   float64      `json:"profit"`
}

type Redistribution struct {
	CryptoFrom map[int]float64 `json:"cryptofrom"`
	CryptoTo map[int]float64 `json:"cryptoto"`
	FiatFrom map[int]float64 `json:"fiatfrom"`
	FiatTo map[int]float64 `json:"fiatto"`
}

type Event struct {
	Time  time.Time  `json:"time"`
	Price float64    `json:"price"`
	Buy   *BuyOrder  `json:"buy"`
	Sell  *SellOrder `json:"sell"`
	AssignProfit *PositionInfo `json:"assignprofit"`
	Redistribution *Redistribution `json:"redistribution"`
	ProfitPercent float64 `json:"profitpercent"`
	TotalAssets float64 `json:"totalassets"`
}

func computeProfitPercent(price float64, engine *TradingEngine) float64{
	allFiat := engine.BankFiat
	allCrypto := engine.BankCrypto
	for _, position := range engine.Positions {
		allCrypto += position.Crypto
		allFiat += position.Fiat
	}

	assetsAfterTrading := allFiat + allCrypto*price
	assetsNoTrading := price * engine.InitialCrypto
	profitPercent := (assetsAfterTrading - assetsNoTrading) * 100.0 / assetsNoTrading
	return profitPercent
}

func computeTotalAssets(price float64, engine *TradingEngine) float64 {
	allFiat := engine.BankFiat
	allCrypto := engine.BankCrypto
	for _, position := range engine.Positions {
		allCrypto += position.Crypto
		allFiat += position.Fiat
	}
	return allFiat + allCrypto*price
}

func NewPriceEvent(time time.Time, price float64, engine *TradingEngine) Event {
	return Event{
		Time:  time,
		Price: price,
		Buy:   nil,
		Sell:  nil,
		AssignProfit: nil,
		ProfitPercent: computeProfitPercent(price, engine),
		TotalAssets: computeTotalAssets(price, engine),
	}
}

func NewBuyEvent(time time.Time, price float64, position *Position, volume float64, currency string, profit float64) Event {
	return Event{
		Time:  time,
		Price: price,
		Buy: &BuyOrder{
			Position: position.AsPositionInfo(),
			Kind:     "buy",
			Volume:   volume,
			Currency: currency,
			Profit:   profit,
		},
		Sell: nil,
		AssignProfit: nil,
		ProfitPercent: computeProfitPercent(price, position.Engine),
		TotalAssets: computeTotalAssets(price, position.Engine),
	}
}

func NewSellEvent(time time.Time, price float64, position *Position, volume float64, currency string) Event {
	return Event{
		Time:  time,
		Price: price,
		Buy:   nil,
		Sell: &SellOrder{
			Position: position.AsPositionInfo(),
			Kind:     "buy",
			Volume:   volume,
			Currency: currency,
		},
		AssignProfit: nil,
		ProfitPercent: computeProfitPercent(price, position.Engine),
		TotalAssets: computeTotalAssets(price, position.Engine),
	}
}

func NewAssignEvent(time time.Time, price float64, position *Position) Event {
	positionInfo := position.AsPositionInfo()
	return Event {
		Time: time,
		Price: price,
		Buy: nil,
		Sell: nil,
		AssignProfit: &positionInfo,
		ProfitPercent:computeProfitPercent(price, position.Engine),
		TotalAssets:computeTotalAssets(price, position.Engine),
	}
}

func NewRedistributeEvent(time time.Time, price float64, redistribution *Redistribution) Event {
	return Event {
		Time: time,
		Price: price,
		Buy: nil,
		Sell: nil,
		AssignProfit: nil,
		ProfitPercent: 0.0,
		Redistribution: redistribution,
	}
}