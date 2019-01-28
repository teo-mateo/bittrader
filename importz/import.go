package importz

import (
	"time"
	"os"
	"log"
	"encoding/json"
	"fmt"
	"github.com/teo-mateo/bittrader/data"
)

type SourceData struct{
	Success bool 		`json:"success"`
	Message string 		`json:"message"`
	Result []PriceItem	`json:"result"`
}

type PriceItem struct{
	O float64 		`json:"O"`
	H float64		`json:"H"`
	L float64		`json:"L"`
	C float64		`json:"C"`
	V float64		`json:"V"`
	T string		`json:"T"`
	BV float64		`json:"BV"`
}

func (pi PriceItem) Time() time.Time{
	t, err := time.Parse("2006-01-02T15:04:05", pi.T)
	if err != nil{
		log.Panic(err)
	}
	return t
}

func ImportData(tradepair_id int, sim_id int, file string){
	f, err := os.Open(file)
	if err != nil{
		log.Panic(err)
	}
	sd := SourceData{}
	dec := json.NewDecoder(f)
	err = dec.Decode(&sd)
	if err != nil{
		log.Panic(err)
	}

	db, _:= data.Connect()
	defer db.Close()

	db.Exec("delete from price1m where sim_id = $1", sim_id)

	for _, pi := range sd.Result{
		fmt.Printf("%.8f on %s\n", pi.C, pi.Time().Format("02/01/2006 03:04:05 PM"))
		query := "insert into price1m (tradepair_id, time, price, imagined, sim_id) values ($1, $2, $3, $4, $5)"
		_, err = db.Exec(query, tradepair_id, pi.Time(), pi.C, true, sim_id)
		if err != nil{
			log.Panic(err)
		}
	}
}

func TestTimeParser(){
	pi := PriceItem{
		T: "2017-10-31T23:13:0",
	}
	t := pi.Time()
	fmt.Println(t)
}