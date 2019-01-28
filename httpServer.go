package main

import (
	"encoding/json"
	"github.com/teo-mateo/bittrader/engine"
	"log"
	"net/http"
	"github.com/teo-mateo/bittrader/broker"
	"github.com/teo-mateo/bittrader/data"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"time"
	"strconv"
	"github.com/teo-mateo/bittrader/bithandlers"
	"github.com/teo-mateo/bittrader/orchestrator"
)

func startHTTPServer() {
	router:= mux.NewRouter()
	router.HandleFunc("/misc/price1msim", func(rw http.ResponseWriter, r *http.Request){
		price1msim, err := data.GetPrice1mSim()
		if err != nil{
			log.Panic(err)
		}

		responsebytes, err := json.Marshal(price1msim)
		if err != nil {
			log.Panic(err)
		}

		rw.Write(responsebytes)
		rw.Header().Set("Content-Type", "application/json")
	})

	router.HandleFunc("/misc/price1msim/{id}/{minuteskip}", func(rw http.ResponseWriter, r *http.Request){
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil{
			log.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		minuteskip, err := strconv.Atoi(vars["minuteskip"])
		if err != nil{
			log.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		data, err := data.GetPrice1mSimData(id, minuteskip)
		if err != nil{
			log.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		databytes, err := json.Marshal(data)
		if err != nil{
			log.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.Write(databytes)

	}).Methods("GET")

	//temporary
	router.HandleFunc("/positions2", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "positions.json")
	})

	router.Handle("/updates", broker.GetServer())

	router.HandleFunc("/misc/price1msim/create", bithandlers.CreatePrice1m).Methods("POST")
	router.HandleFunc("/misc/price1msim/duplicate", bithandlers.DuplicatePrice1m).Methods("POST")
	router.HandleFunc("/misc/price1msim/delete", bithandlers.DeletePrice1mSim).Methods("POST")
	router.HandleFunc("/misc/tradepairs", bithandlers.GetTradePairs).Methods("GET")

	router.HandleFunc("/engine/stop", func(rw http.ResponseWriter, r *http.Request){
		orchestrator.Stop()
	}).Methods("POST")
	router.HandleFunc("/engine/pause", func(rw http.ResponseWriter, r *http.Request){
		orchestrator.Pause()
	}).Methods("POST")
	router.HandleFunc("/engine/play", func(rw http.ResponseWriter, r *http.Request){
		orchestrator.Play()
	}).Methods("POST")
	router.HandleFunc("/engine/play/new", func(rw http.ResponseWriter, r *http.Request){

	}).Methods("POST")

	staticHandler := bithandlers.StaticFilesHandler{
		Prefix:"bittrader/",
	}
	router.PathPrefix("/bittrader/").Handler(staticHandler)

	e := orchestrator.GetEngine()
	if e != nil {
		router.HandleFunc("/positions", func(rw http.ResponseWriter, r *http.Request) {
			var positions []engine.PositionInfo = make([]engine.PositionInfo, 0, 0)
			for i := len(e.Positions) - 1; i >= 0; i-- {
				positions = append(positions, e.Positions[i].AsPositionInfo())
			}

			body, err := json.Marshal(positions)
			if err != nil {
				log.Fatal(err)
			}

			rw.Header().Add("Content-Type", "application/json")
			rw.Write(body)

		})
	}

	go func() {

		headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
		originsOk := handlers.AllowedOrigins([]string{"*"})
		methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})



		localserver := "localhost:3000"
		srv := http.Server{
			Addr:localserver,
			Handler:handlers.CORS(headersOk, originsOk, methodsOk)(router),
			WriteTimeout:15*time.Second,
			ReadTimeout:15*time.Second,
		}

		fmt.Println("Start HTTP server. ", localserver)
		log.Fatal(srv.ListenAndServe())
	}()
}
