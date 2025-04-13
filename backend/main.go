package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Obj1 struct {
	Day  string `json:"day"`
	Time string `json:"time"`
	Task string `json:"task"`
}

var db *sql.DB

func init() { //runs when program strats before main once
	connstr := "postgresql://postgres:qwer@localhost:5432/test?sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connstr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("database connected sucssefully")

}
func main() {

	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/task", enableCORS(changeDb))
	mux.HandleFunc("/reset", enableCORS(resetTable))
	mux.HandleFunc("/schedule", enableCORS(getSchedule))

	fmt.Println("server satrted at port 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

//function to get the ip of user

func getTable(w http.ResponseWriter, r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	fmt.Println("ip fetching sucessfull :", ip)

	//error with nat gateways
	//experimetal code from sonet

	cookie, err := r.Cookie("device_id") //  get cokkires form browser
	if err != nil {                      
		// Create new cookie 
		deviceID := uuid.New().String() // Generate unique ID: "abc123..."
		cookie = &http.Cookie{
			Name:     "device_id",                          // Label: "device_id"
			Value:    deviceID,                             // Content: "abc123..."
			Expires:  time.Now().Add( 24 * time.Hour),      // Valid for 1 day
			Path:     "/",                                  // Valid everywhere on site
			HttpOnly: true,                                 // Browser only, no JavaScript
		}
		http.SetCookie(w, cookie) //  "Here browser, store this cookie!"
	} else {
		// Browser: "Yes, here's your cookie!" 
		fmt.Println("Got existing cookie:", cookie.Value)
	}

	//experimental code ends

	safeIp := strings.ReplaceAll(ip, ".", "_")
	tableName := fmt.Sprintf("schedule_%s_%s", safeIp, cookie.Value[:8])
	query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            id SERIAL PRIMARY KEY,
            day VARCHAR(3) UNIQUE,
            nine_am TEXT DEFAULT '',
            ten_am TEXT DEFAULT '',
            eleven_am TEXT DEFAULT '',
            twelve_pm TEXT DEFAULT '',
            one_pm TEXT DEFAULT '',
            two_pm TEXT DEFAULT '',
            three_pm TEXT DEFAULT '',
            four_pm TEXT DEFAULT ''
        );
    `, tableName)

	_, err = db.Exec(query)
	if err != nil {
		fmt.Println("table creation error")
		http.Error(w, "table creation error", http.StatusInternalServerError)
		return "" //we are retuning a empty string here
	}

	//now check if the rows are zero for given table and then add the days if rows are indeed zero
	var count int
	countQuery := fmt.Sprintf("select count(*) from %s", tableName)

	err = db.QueryRow(countQuery).Scan(&count)
	if err != nil {
		fmt.Println("error while fetching coloums")
		return ""
	}
	if count == 0 {
		days := []string{"mon", "tus", "wed", "thu", "fri", "sat", "sun"}

		for _, v := range days {
			dayInsertQuery := fmt.Sprintf("insert into %s(day) values ($1)", tableName)
			_, err = db.Exec(dayInsertQuery, v)
			if err != nil {
				fmt.Println("day insertition error")
				http.Error(w, "day insertiton error", http.StatusInternalServerError)
				return ""
			}
		}
	}
	return tableName
}

func changeDb(w http.ResponseWriter, r *http.Request) {
	tableName := getTable(w, r)
	if r.Method != "POST" {
		http.Error(w, "wrong method used", http.StatusMethodNotAllowed)
		return
	}
	//storing json
	var work Obj1
	err := json.NewDecoder(r.Body).Decode(&work)
	if err != nil {
		http.Error(w, "Invalid json Format", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf("update %s set %s=$1 where day=$2;", tableName, work.Time)

	_, err2 := db.Exec(query, work.Task, work.Day)
	if err2 != nil {
		http.Error(w, "databse error", http.StatusInternalServerError)
		fmt.Println(err)
		log.Print()
	}
	fmt.Println("done updating the task")

}

func resetTable(w http.ResponseWriter, r *http.Request) {
	tableName := getTable(w, r)
	fmt.Println("reset started")
	if r.Method != "POST" {
		http.Error(w, "wrong method used", http.StatusMethodNotAllowed)
		return
	}
	query := fmt.Sprintf(`UPDATE %s SET 
		nine_am='', ten_am='', eleven_am='', twelve_pm='',
		one_pm='', two_pm='', three_pm='', four_pm=''`, tableName)
	_, err := db.Exec(query)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	fmt.Println("table reset completed")

}

type Fetcher struct {
	ID       int    `json:"id"`
	Day      string `json:"day"`
	NineAM   string `json:"nine_am"`
	TenAM    string `json:"ten_am"`
	ElevenAM string `json:"eleven_am"`
	TwelvePM string `json:"twelve_pm"`
	OnePM    string `json:"one_pm"`
	TwoPM    string `json:"two_pm"`
	ThreePM  string `json:"three_pm"`
	FourPM   string `json:"four_pm"`
}

// to fetch schedule when website loads
func getSchedule(w http.ResponseWriter, r *http.Request) {
	fmt.Println("sending started")
	tableName := getTable(w, r)
	query := fmt.Sprintf(`select * from %s`, tableName)
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	defer rows.Close() //dont forgot to close rows coonection
	var schedules []Fetcher
	var schedule Fetcher
	for rows.Next() {
		err = rows.Scan(&schedule.ID, &schedule.Day, &schedule.NineAM, &schedule.TenAM, &schedule.ElevenAM, &schedule.TwelvePM, &schedule.OnePM, &schedule.TwoPM, &schedule.ThreePM, &schedule.FourPM)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Error reading data", http.StatusInternalServerError)
			return
		}
		schedules = append(schedules, schedule)
	}

	//sending json back to user
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(schedules)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error encoding data", http.StatusInternalServerError)
		return
	}
	fmt.Println("sucessfully send json schedules")

}

//here we are doing something intresting we are adding middle ware, evertime browser sends a request cors req needs to be executed so we are making cors execute before main function

func enableCORS(originalHandler http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return //return from further execution
		}

		originalHandler(w, r) //run the original function as usual
	}
}
