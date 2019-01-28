package tradesapi

import (
	"github.com/beldur/kraken-go-api-client"
	"os"
)

func GetTrades(pair string, since int64) (*krakenapi.TradesResponse, error) {
	key := os.Getenv("KRAPI_KEY_1")
	secret := os.Getenv("KRAPI_SECRET_1")
	kapi := krakenapi.New(key, secret)

	trades, err := kapi.Trades(pair, since)
	if err != nil {
		return nil, err
	}

	return trades, nil
}
