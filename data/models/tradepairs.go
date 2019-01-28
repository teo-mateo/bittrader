package models

import (
	"time"
)

type Tradepair struct {
	Id         int
	Cfrom      string
	Cto        string
	Nicename   string
	Krakenname string
	Krakenlast int64
}

type Simulation struct {
	Id               int64
	Tradepair        Tradepair
	Start            time.Time
	End              time.Time
	Investment       float64
	RangeBottom      float64
	RangeTop         float64
	MaxPositions     float64
	TrendRate        float64
	PositionStepover float64
	Time             time.Time
	StartPrice       float64
	EndPrice         float64
	Performance      float64
	Positions        []*Position
}

type Position struct {
	Id         int64
	Simulation *Simulation
	Top        float64
	Bottom     float64
}
