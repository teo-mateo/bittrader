package myapi

import (
	"github.com/beldur/kraken-go-api-client"
	"github.com/teo-mateo/bittrader/data"
	"log"
	"os"
	"time"
	"github.com/davecgh/go-spew/spew"
	"reflect"
	"fmt"
)

var Cache map[int]map[time.Time]float64 = make(map[int]map[time.Time]float64)

func GetPriceAtTime(sim_id int, t time.Time) float64 {

	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
	var cache map[time.Time]float64

	if _, ok := Cache[sim_id]; ok {
		cache = Cache[sim_id]
	} else {

		cache, err := data.PreloadTradeData(sim_id)
		if err != nil {
			log.Panic(err)
		}
		Cache[sim_id] = cache
	}

	//now := time.Now()
	//today := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.Local)
	//if t.After(today){
	//
	//	pair, err := data.GetTradePair(tradepair_id)
	//	if err != nil{
	//		log.Fatal(err)
	//	}
	//
	//	key := os.Getenv("KRAPI_KEY_1")
	//	secret := os.Getenv("KRAPI_SECRET_1")
	//	kapi := krakenapi.New(key, secret)
	//	response, err := kapi.Ticker(pair.Krakenname)
	//	if err != nil{
	//		log.Fatal(err)
	//	}
	//
	//	getTickerField(response, pair.Krakenname)
	//}
	cache = Cache[sim_id]
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
	price := cache[t]
	return price
}

func getTickerField(ticker *krakenapi.TickerResponse, pair string) float64 {

	if pair == "XETHZEUR" {
		return float64(ticker.XETHZEUR.OpeningPrice)
	}

	return -1.0
}


type AssetPair struct{
	Name string
	Info *krakenapi.AssetPairInfo
	MapToTradePair int
}

func GetAssetPairs() (map[string]krakenapi.AssetPairInfo, error) {
	fmt.Println()

	key := os.Getenv("KRAPI_KEY_1")
	secret := os.Getenv("KRAPI_SECRET_1")
	kapi := krakenapi.New(key, secret)
	var response *krakenapi.AssetPairsResponse
	response, err := kapi.AssetPairs()
	if err != nil {
		log.Fatal(err)
	}
	spew.Config.MaxDepth = 10
	spew.Dump(response.XETHZEUR)


	r := make(map[string]krakenapi.AssetPairInfo, 0)

	t := reflect.TypeOf(*response)
	for i := 0; i < t.NumField(); i++{
		name := t.Field(i).Name
		val := reflect.ValueOf(*response).FieldByName(name)
		r[name] = val.Interface().(krakenapi.AssetPairInfo)
	}

	//db, err := data.Connect()
	//if err != nil{
	//	return nil, err
	//}

	return r, nil

}
