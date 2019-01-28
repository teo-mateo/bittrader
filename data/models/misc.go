package models

import "time"

/*
id
name
description
tradepair_id
*/

type Price1mSim struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
	TradePairId int `json:"tradepair_id"`
	MinTime time.Time `json:"mintime"`
	MaxTime time.Time `json:"maxtime"`
	Tradepair Tradepair
}