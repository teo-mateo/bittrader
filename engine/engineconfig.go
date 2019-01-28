package engine

import "time"

const COMMISSION_PC = 0.0026

var POSITION_INCREMENT_PC float64 = 2.0

var MOVE_PROFIT_ABOVE_PC float64 = 0
var POSITION_STEPOVER float64 = 0.0
var USE_ONLY_PROFITS bool = false
var SAVE_TRADES = false
var TIME_INCREMENT = 1 * time.Minute
var OP_SLEEP = 3 * time.Millisecond

var MIN_PROFIT_FIAT float64 = 0.0
//var BROKER_ONLINE = true

var MONTHLY_BUYIN = 0.0

var MIDDLE_PC = 50.0 // [-10% .. 10%]
var PLUCK_PC = 20.0