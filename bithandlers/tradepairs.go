package bithandlers

import (
	"net/http"
	"github.com/teo-mateo/bittrader/data"
	"log"
	"encoding/json"
)

func GetTradePairs(rw http.ResponseWriter, r *http.Request) {

	//rw.Header().Set("Access-Control-Allow-Origin", "*")
	//rw.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	db, err := data.Connect()
	if err != nil{
		sendHttpError(rw, err)
		return
	}

	rows, err := db.Query("select id, cfrom, cto, nicename, krakenname from tradepairs order by id asc")
	if err != nil{
		sendHttpError(rw, err)
		return
	}

	type Row struct{
		Id 			int 	`json:"id"`
		Cfrom 		string	`json:"cfrom"`
		Cto 		string	`json:"cto"`
		Nicename 	string	`json:"nicename"`
		Krakenname 	string	`json:krakenname`
	}
	result := make([]Row, 0)
	for rows.Next(){
		row := Row{}
		err = rows.Scan(&row.Id, &row.Cfrom, &row.Cto, &row.Nicename, &row.Krakenname)
		if err != nil{
			sendHttpError(rw, err)
			return
		}

		result = append(result, row)
	}

	bytes, err := json.Marshal(result)
	if err != nil{
		sendHttpError(rw, err)
		return
	}

	rw.Write(bytes)
	return

}

func sendHttpError(rw http.ResponseWriter, err error){
	log.Println(err)
	rw.WriteHeader(http.StatusInternalServerError)
	rw.Write([]byte(err.Error()))
}