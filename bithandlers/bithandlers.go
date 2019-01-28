package bithandlers

import (
	"github.com/Sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"log"
	"time"
	"encoding/json"
	"github.com/teo-mateo/bittrader/data"
	"errors"
	"fmt"
	"github.com/teo-mateo/bittrader/api"
)


type StaticFilesHandler struct {
	Prefix string
}

//StaticFilesHandler ...
func (h StaticFilesHandler ) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	//get the url of the file to serve
	src := req.RequestURI[1:]
	src = strings.Replace(src, h.Prefix, "", 1)

	//file should be under current dir
	cwd, err := os.Getwd()
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	//build file name
	fname := filepath.Join(cwd, src)
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	logrus.Info(fname)

	//serve file
	http.ServeFile(rw, req, fname)
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "static/", http.StatusMovedPermanently)
}


type CreateDummyPrices struct {
	SimId int `json:"simid"`
	StartTime *time.Time `json:"starttime"`
	EndTime *time.Time `json:"endtime"`
	NewSimName string `json:"newsimname"`
	NewSimDescription string `json:"newsimdesc"`
}

func DeletePrice1mSim(rw http.ResponseWriter, req *http.Request){
	decoder := json.NewDecoder(req.Body)
	var id int
	err := decoder.Decode(&id)
	if err != nil{
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	err = data.DeletePrice1mSim(id)
	if err != nil{
		if err != nil{
			log.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}
	}
}

func DuplicatePrice1m(rw http.ResponseWriter, req *http.Request){

	cdp := CreateDummyPrices{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&cdp)
	if err != nil{
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	err = data.DuplicatePrice1m(cdp.SimId, cdp.NewSimName, cdp.NewSimDescription)
	if err != nil{
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}
}

func CreatePrice1m(rw http.ResponseWriter, req *http.Request){

	cdp := CreateDummyPrices{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&cdp)
	if err != nil{
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	if cdp.StartTime == nil || cdp.EndTime == nil{
		err = errors.New("missing starttime or endtime")
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	err = CreatePrice1m2(cdp.SimId, *cdp.StartTime, *cdp.EndTime)
	if err != nil{

		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

}

//CreatePrice1m will append new prices based on the (start->end) selection.
func CreatePrice1m2(sim_id int, start time.Time, end time.Time) error {

	fmt.Println("Creating new prices")

	//db connect
	db, err := data.Connect()
	if err != nil{
		return err
	}
	defer db.Close()

	//get tradepair_id
	var tradepair_id int
	row := db.QueryRow("select tradepair_id from price1msim where id = $1", sim_id)
	err = row.Scan(&tradepair_id)
	if err != nil{
		return err
	}

	//get last price time
	lpt, err := data.GetLastPriceTime(db, sim_id)
	if err != nil{
		return err
	}
	//get price at last time
	priceAtLpt := myapi.GetPriceAtTime(sim_id, *lpt)
	//get price at start time
	priceAtStart := myapi.GetPriceAtTime(sim_id, start)
	//compute delta
	delta := priceAtLpt - priceAtStart
	timedelta := end.Sub(start)
	//loop (start -> end)
	t:= lpt.Add(time.Minute)

	minutes := int(timedelta.Minutes())
	for i:= 1; i < minutes; i++{

		//price:= myapi.GetPriceAtTime(sim_id, start.Add(time.Duration(i)*time.Minute))
		price:= myapi.GetPriceAtTime(sim_id, start.Add(time.Duration(i)*time.Minute))
		price += delta

		_, err := db.Exec("INSERT INTO public.price1m(tradepair_id, \"time\", price, imagined, sim_id) VALUES ($1, $2, $3, $4, $5)", tradepair_id, t, price, true, sim_id)
		if err != nil {
			return err
		}

		if i % 60 == 0{
			fmt.Printf("new price @ %v, %.3f\n", t, price)
		}

		t = t.Add(time.Minute)
	}

	query := `update price1msim s set
		starttime = (select min(p.time) from price1m p where p.sim_id = s.id),
		endtime = (select max(p.time) from price1m p where p.sim_id = s.id)
		where s.id = $1`

	_, err = db.Exec(query, sim_id)
	if err != nil{
		return err
	}


	if err != nil{
		return err
	}

	return nil
}