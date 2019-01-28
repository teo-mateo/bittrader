package data

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/beldur/kraken-go-api-client"
	_ "github.com/lib/pq"
	"github.com/teo-mateo/bittrader/data/models"
	"github.com/teo-mateo/bittrader/tradesapi"
	"log"
	"os"
	"time"
)

func Connect() (db *sql.DB, err error) {
	psqlInfo := pgConnectionString()
	db, err = sql.Open("postgres", psqlInfo)
	return
}

// PgConnectionString ...
func pgConnectionString() string {
	return os.Getenv("PG_CN_TRADING")
}

func InsertTrade(tradepair_id int, price float64, volume float64, time float64, buysell rune, marketlimit rune) error {
	_, err := Connect()
	if err != nil {
		return err
	}

	query := fmt.Sprintf("insert into trades (tradepair_id, price, volume, time, buysell, marketlimit) values (%d, %d, %d, %d, '%s', '%s')",
		tradepair_id, price, volume, time, buysell, marketlimit)
	fmt.Println(query)

	return nil
}

func UpdatePair(tradepair_id int, kraken_last int64) error {
	db, err := Connect()
	if err != nil {
		return err
	}

	defer db.Close()

	query := fmt.Sprintf("update tradepairs set krakenlast = %d where id = %d", kraken_last, tradepair_id)
	fmt.Println(query)

	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func PreloadTradeData(sim_id int) (map[time.Time]float64, error) {

	fmt.Printf("Preloading prices for sim: %d\n", sim_id)
	db, err := Connect()
	if err != nil {
		return nil, err
	}

	defer db.Close()

	rows, err := db.Query("select time, price from price1m where sim_id = $1 order by time asc", sim_id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make(map[time.Time]float64)
	for rows.Next() {
		t := time.Now()
		p := 0.0
		rows.Scan(&t, &p)
		t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
		result[t] = p
	}

	fmt.Printf("Preloaded %d prices.\n\n", len(result))
	time.Sleep(time.Second)
	return result, nil
}

func GetTradePairs() (tradepairs []models.Tradepair, err error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := "select * from tradepairs"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]models.Tradepair, 0)
	for rows.Next() {
		pair := models.Tradepair{}
		err := rows.Scan(&pair.Id, &pair.Cfrom, &pair.Cto, &pair.Nicename, &pair.Krakenname, &pair.Krakenlast)
		if err != nil {
			return nil, err
		}

		result = append(result, pair)
	}

	return result, nil
}

func GetSim(sim_id int) (models.Price1mSim, error){
	db, err := Connect()
	if err != nil {
		return models.Price1mSim{}, err
	}
	defer db.Close()

	rows, err := db.Query("select id, name, description, tradepair_id from price1msim where id = $1", sim_id)
	if err != nil {
		return models.Price1mSim{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return models.Price1mSim{}, errors.New(fmt.Sprintf("no tradepair found for id: %d", sim_id))
	}

	sim := models.Price1mSim{}
	err = rows.Scan(&sim.Id, &sim.Name, &sim.Description, &sim.TradePairId)
	if err != nil {
		return models.Price1mSim{}, err
	}

	sim.Tradepair, err = GetTradePair(sim.TradePairId)
	if err != nil {
		return models.Price1mSim{}, err
	}

	return sim, nil
}

func GetTradePair(tradepair_id int) (models.Tradepair, error) {
	db, err := Connect()
	if err != nil {
		return models.Tradepair{}, err
	}
	defer db.Close()

	rows, err := db.Query("select id, cfrom, cto, nicename, krakenname, krakenlast from tradepairs where id = $1", tradepair_id)
	if err != nil {
		return models.Tradepair{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return models.Tradepair{}, errors.New(fmt.Sprintf("no tradepair found for id: %d", tradepair_id))
	}

	pair := models.Tradepair{}
	err = rows.Scan(&pair.Id, &pair.Cfrom, &pair.Cto, &pair.Nicename, &pair.Krakenname, &pair.Krakenlast)
	if err != nil {
		return models.Tradepair{}, err
	}
	return pair, nil
}

func SaveTrade(tx *sql.Tx, info krakenapi.TradeInfo, tradepair_id int) (*time.Time, error) {
	buysell := "b"
	if info.Sell {
		buysell = "s"
	}
	marketlimit := "m"
	if info.Limit {
		marketlimit = "l"
	}

	tradetime := time.Unix(info.Time, 0)

	_, err := tx.Exec("INSERT INTO trades(tradepair_id, price, volume, time, buysell, marketlimit) VALUES ($1, $2, $3, $4, $5, $6)",
		tradepair_id, info.PriceFloat, info.VolumeFloat, tradetime, buysell, marketlimit)
	if err != nil {
		return nil, err
	}

	return &tradetime, nil
}

func UpdateTrades(sim_id int) (*time.Time, error) {

	sim, err := GetSim(sim_id)
	if err != nil{
		return nil, err
	}

	pair, err := GetTradePair(sim.TradePairId)
	if err != nil {
		return nil, err
	}

	fmt.Println("before update:")
	fmt.Println(pair)

	trades, err := tradesapi.GetTrades(pair.Krakenname, pair.Krakenlast)
	if err != nil {
		return nil, err
	}

	db, err := Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	t := time.Now()
	lastTradeTime := &t

	tx, err := db.Begin()
	for _, ti := range trades.Trades {
		//fmt.Printf("Updating: %v\n", ti)
		lastTradeTime, err = SaveTrade(tx, ti, sim.TradePairId)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	query := fmt.Sprintf("update tradepairs set krakenlast = %d where id = %d", trades.Last, sim.TradePairId)
	fmt.Println(query)

	_, err = tx.Exec(query)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return lastTradeTime, nil
}

func GenerateDummyPrice1mForInterval(sim_id int, newsimname string, start time.Time, end time.Time) error {
	db, err := Connect()
	if err != nil{
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil{
		return err
	}

	t, err := GetLastPriceTime(db, sim_id)
	if err != nil{
		return err
	}

	end = t.Add(end.Sub(start))

	for {
		if t.After(end){
			err = tx.Commit()
			if err != nil{
				err = tx.Rollback()
				return err
			}
			break
		}

		t2 := t.Add(1 * time.Minute)
		t = &t2

		//get price at t


	}

	return nil
}

func GetLastPriceTime(db *sql.DB, sim_id int) (*time.Time, error) {
	//get last price time
	rows, err := db.Query("select \"time\" from price1m where sim_id = $1 order by time desc limit 1", sim_id)
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()

	lpt := time.Now()
	if rows.Next() {
		err = rows.Scan(&lpt)
		fmt.Println("Last price time: ", lpt)
		if err != nil {
			return nil, err
		}
	} else {
		log.Panic("couldn't find any last time.")
	}

	return &lpt, nil
}

func GenerateDummyPrice1m(sim_id int, days int64, growth_pc int64) {
	db, err := Connect()
	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	sim, err := GetSim(sim_id)
	if err != nil{
		log.Panic(err)
	}

	lpt, err := GetLastPriceTime(db, sim_id)
	if err != nil{
		log.Panic(err)
	}

	var m int64
	minutes := days * 24 * 60

	growth_pm := float64(growth_pc / minutes)

	for m = 1; m <= minutes; m++ {
		tpast := lpt.Add(-1 * time.Duration(m) * time.Minute)
		price := getPriceAtTime(db, sim_id, tpast)

		//adjust price
		price = price + price * growth_pm / 100

		tfuture := lpt.Add(time.Duration(m) * time.Minute)
		if (m % 120) == 0 {
			fmt.Printf("At %v it will be %f, as it was on %v\n", tfuture, price, tpast)
		}
		_, err = db.Exec("INSERT INTO public.price1m(tradepair_id, \"time\", price, imagined, sim_id) VALUES ($1, $2, $3, true, $4)", sim.TradePairId, tfuture, price, sim_id)
		if err != nil {
			log.Panic(err)
		}
	}
}



func DuplicatePrice1m(sim_id int, newname string, newdescription string) error {

	fmt.Println("Duplicating price1m set")

	db, err := Connect()
	if err != nil{
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil{
		return err
	}

	//create new sim
	query := `
	insert into price1msim (name, description, tradepair_id)
	select $2, $3, p2.tradepair_id
	from price1msim p2 where p2.id = $1 returning id`

	row := tx.QueryRow(query, sim_id, newname, newdescription)
	var newsim_id int
	err = row.Scan(&newsim_id)
	if err != nil{
		tx.Rollback()
		return err
	}

	query = `insert into price1m (tradepair_id, time, price, imagined, sim_id)
		select p2.tradepair_id, p2.time, p2.price, true, $1
		from price1m p2 where p2.sim_id = $2`

	_, err = tx.Exec(query, newsim_id, sim_id)
	if err != nil{
		tx.Rollback()
		return err
	}

	query = `update price1msim s set
		starttime = (select min(p.time) from price1m p where p.sim_id = s.id),
		endtime = (select max(p.time) from price1m p where p.sim_id = s.id)
		where s.id = $1`

	_, err = tx.Exec(query, newsim_id)
	if err != nil{
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func DeletePrice1mSim(id int) error{
	db, err := Connect()
	if err != nil{
		return err
	}

	defer db.Close()

	tx, err := db.Begin()
	if err != nil{
		return err
	}


	_, err = tx.Exec("delete from price1m where sim_id = $1", id)
	if err != nil{
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("delete from price1msim where id = $1", id)
	if err != nil{
		tx.Rollback()
		return err
	}



	err = tx.Commit()
	if err != nil{
		return err
	}

	return nil
}

func GeneratePrice1h(tradepair_id int) {
	db, err := Connect()
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	//get first, last price time
	row := db.QueryRow("select min(\"time\") as min, max(\"time\") as max from price1m where tradepair_id = $1", tradepair_id)
	var first time.Time
	var last time.Time
	err = row.Scan(&first, &last)
	if err != nil {
		log.Panic(err)
	}

	first = time.Date(first.Year(), first.Month(), first.Day(), first.Hour(), 0, 0, 0, time.UTC)
	first = first.Add(time.Hour)
	last = time.Date(last.Year(), last.Month(), last.Day(), last.Hour(), 0, 0, 0, time.Local)
	last = last.Add((-1) * time.Hour)

	fmt.Printf("first price1m time of %d: %v\n", tradepair_id, first)
	fmt.Printf("last price1m time of %d: %v\n", tradepair_id, last)

	tx, err := db.Begin()
	if err != nil {
		log.Panic(err)
	}

	//cleanup
	_, err = tx.Exec("delete from price1h where tradepair_id = $1", tradepair_id)
	if err != nil {
		log.Panic(err)
	}

	for t := first; t.Before(last); t = t.Add(time.Hour) {
		hourstart := t
		hourend := t.Add(59 * time.Minute)

		row := tx.QueryRow("select min(x.price), max(x.price), avg(x.price), median(x.price), stddev(x.price), count(1) from (select * from price1m where tradepair_id = $1 and \"time\" >= $2 and \"time\" < $3 order by time) x", tradepair_id, hourstart, hourend)

		var min, max, avg, median, stddev, count float64
		err = row.Scan(&min, &max, &avg, &median, &stddev, &count)
		if err != nil {
			//log.Panic(err)
			continue
		}

		stdoveravg := 100.0 * stddev / avg

		fmt.Printf("%s -> MIN %.3f MAX %.3f AVG %.3f MED %.3f STDDEV %.3f STDOVERAVG %.3f\n", t, min, max, avg, median, stddev, stdoveravg)

		_, err = tx.Exec("insert into price1h (tradepair_id, time, price, low, high, average, median, stddev, stdoveravg, imagined) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)", tradepair_id, t, avg, min, max, avg, median, stddev, stdoveravg, false)
		if err != nil {
			log.Panic(err)
		}
	}

	tx.Commit()
}

func GeneratePrice1m(sim_id int) {
	db, err := Connect()
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	sim, err := GetSim(sim_id)
	if err != nil{
		log.Panic(err)
	}

	//get last price time
	rows, err := db.Query("select \"time\" from price1m where sim_id = $1 order by time desc limit 1", sim_id)
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()

	t := time.Now()
	fmt.Println(t)
	resume := false

	if rows.Next() {
		err = rows.Scan(&t)
		if err != nil {
			log.Panic(err)
		}
		resume = true
	}

	if !resume {
		//get first time from trades
		rows, err := db.Query("select \"time\" from trades where tradepair_id = $1 order by time asc limit 1", sim.TradePairId)
		if err != nil {
			log.Panic(err)
		}
		defer rows.Close()
		if rows.Next() {
			err = rows.Scan(&t)
			fmt.Println(t)
			if err != nil {
				log.Panic(err)
			}
		} else {
			log.Panic(errors.New("no trade"))
		}
	}

	cutoff := time.Now()
	rows, err = db.Query("select \"time\" from trades where tradepair_id = $1 order by time desc limit 1", sim.TradePairId)
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&cutoff)
		fmt.Println(cutoff)
		if err != nil {
			log.Panic(err)
		}
	} else {
		log.Panic(errors.New("no trade"))
	}

	//starting at this time
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)
	fmt.Println(t)

	for {
		//cutoff check
		if t.After(cutoff) {
			break
		}

		//inc time
		t = t.Add(time.Minute)

		//get price and save
		price := getTradedPriceAtTime(db, sim.TradePairId, t)
		_, err := db.Exec("INSERT INTO public.price1m(tradepair_id, \"time\", price, sim_id) VALUES ($1, $2, $3, $4)", sim.TradePairId, t, price, sim.Id)
		if err != nil {
			log.Panic(err)
		}

		//say something
		fmt.Printf("At %s, price of %d was %f\n", t.Format("Mon Jan _2 15:04:05 2006"), sim.TradePairId, price)

	}

}

func getPriceAtTime(db *sql.DB, sim_id int, t time.Time) float64 {
	var price float64 = -1.0
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)
	rows, err := db.Query("select price from price1m where sim_id = $1 and time = $2 limit 1", sim_id, t)
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()
	if !rows.Next() {
		return price
	}
	err = rows.Scan(&price)
	if err != nil {
		log.Panic(err)
	}

	return price
}

func GetPriceAtTime(tradepair_id int, t time.Time) float64 {
	db, err := Connect()
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	return getPriceAtTime(db, tradepair_id, t)
}

func getTradedPriceAtTime(db *sql.DB, tradepair_id int, t time.Time) float64 {
	rows, err := db.Query("select price from trades where tradepair_id = $1 and time <= $2 order by time desc limit 1", tradepair_id, t)
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()

	var price float64
	if !rows.Next() {
		log.Panic(errors.New("no price"))
	}
	err = rows.Scan(&price)
	if err != nil {
		log.Panic(err)
	}
	return price
}

func SaveNewSimulation(simulation *models.Simulation) error {
	db, err := Connect()
	if err != nil {
		return err
	}

	defer db.Close()

	var simulationId int64 = 0

	row := db.QueryRow("insert into simulation (tradepair_id, \"start\", \"end\", investment, range_bottom, max_positions, MOVE_PROFIT_ABOVE_PC, position_stepover) values ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		simulation.Tradepair.Id, simulation.Start, simulation.End, simulation.Investment, simulation.RangeBottom, simulation.MaxPositions, simulation.TrendRate, simulation.PositionStepover)
	err = row.Scan(&simulationId)

	if err != nil {
		return err
	}

	//set the id on the model
	simulation.Id = simulationId

	return nil
}

func UpdateSimulation(id int64, column string, value interface{}) error {
	db, err := Connect()
	if err != nil {
		return err
	}

	defer db.Close()

	query := fmt.Sprintf("update simulation set %s = $1 where id = $2", column)
	_, err = db.Exec(query, value, id)
	if err != nil {
		return err
	}

	return nil
}

func SaveNewPosition(simulation *models.Simulation, top float64, bottom float64) (*models.Position, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var positionId int64 = 0
	row := db.QueryRow("insert into positions (simulation_id, top, bottom) values ($1, $2, $3) RETURNING id", simulation.Id, top, bottom)
	err = row.Scan(&positionId)

	if err != nil {
		return nil, err
	}

	position := models.Position{
		Id:         positionId,
		Top:        top,
		Bottom:     bottom,
		Simulation: simulation,
	}

	return &position, nil
}

func PositionSell(positionId int64, time time.Time, price float64, fiat float64) error {

	db, err := Connect()
	if err != nil {
		return err
	}

	defer db.Close()

	// new record in positiontrades
	_, err = db.Exec("insert into positiontrades (position_id, buysell, \"time\", price, volumecrypto, volumefiat) values ($1, $2, $3, $4, $5, $6)",
		positionId, "s", time, price, fiat/price, fiat)
	if err != nil {
		return err
	}

	// update position w/ how much fiat will be in this position and null for crypto
	_, err = db.Exec("update positions set crypto = NULL, fiat = $1 where id = $2", fiat, positionId)
	if err != nil {
		return err
	}

	// no error
	return nil
}

func PositionBuy(positionId int64, time time.Time, price float64, crypto float64, profit float64) error {

	db, err := Connect()
	if err != nil {
		return err
	}

	defer db.Close()

	// new record in positiontrades
	_, err = db.Exec("insert into positiontrades (position_id, buysell, \"time\", price, volumecrypto, volumefiat, profit) values ($1, $2, $3, $4, $5, $6, $7)",
		positionId, "b", time, price, crypto, crypto*price, profit)
	if err != nil {
		return err
	}

	// update position w/ how much crypto will be in this position and null for fiat
	_, err = db.Exec("update positions set crypto = $1, totalprofit=totalprofit+$2, hitcount=hitcount+1, fiat = NULL where id = $3", crypto, profit, positionId)
	if err != nil {
		return err
	}

	// no error
	return nil
}

func GetPrice1mSim() ([]models.Price1mSim, error){
	db, err := Connect()
	if err != nil {
		return nil, err
	}

	defer db.Close()

	result := make([]models.Price1mSim, 0)

	query := `
		select
			ps.id, ps.name, ps.description, ps.tradepair_id, ps.starttime as mintime, ps.endtime as maxtime
		from price1msim ps order by id asc`
	rows, err := db.Query(query)
	if err != nil{
		return nil, err
	}

	for rows.Next(){
		p1ms := models.Price1mSim{}
		err := rows.Scan(&p1ms.Id, &p1ms.Name, &p1ms.Description, &p1ms.TradePairId, &p1ms.MinTime, &p1ms.MaxTime)
		if err != nil{
			return nil, err
		}
		result = append(result, p1ms)
	}
	return result, nil
}

func GetPrice1mSimData(id int, minuteskip int) ([]interface{}, error) {
	db, err := Connect()
	if err != nil{
		return nil, err
	}

	defer db.Close()

	result := make([]interface{}, 0)

	query := `
		select t.time, t.price
		from (
			select time, price, row_number() over (order by id asc) as row
			from price1m where sim_id = $1
		) t
	where t.row % $2 = 0 order by t.time asc`

	rows, err := db.Query(query, id, minuteskip)
	if err != nil{
		return nil, err
	}

	var r  struct{
		Time time.Time `json:"time"`
		Price float64 `json:"price"`
	}

	for rows.Next(){
		err = rows.Scan(&r.Time, &r.Price)
		if err != nil{
			return nil, err
		}
		result = append(result, r)
	}

	return result, nil
}

