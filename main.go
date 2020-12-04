package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	host   = "10.0.0.22"
	port   = 5432
	user   = "pi"
	dbname = "touchSource"
)

var password string

func main() {

	var pwFlag = flag.String("pw", "", "please include a -pw flag for your database password")
	flag.Parse()
	if *pwFlag == "" {
		fmt.Println("please include a -pw flag for your database password")
		return
	}
	password = *pwFlag

	fmt.Println("Listening on port 18080...")
	http.HandleFunc("/", homePage)
	http.HandleFunc("/people", getPeople)
	// http.HandleFunc("/person/{first}/{last}", getPerson)
	log.Fatal(http.ListenAndServe(":18080", nil))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
	fmt.Println("homePage")
}

func getPerson(w http.ResponseWriter, r *http.Request) {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	rows, err := db.Query("SELECT id, first, last from people order by first desc limit 20")
	if err != nil {
		panic(err)
	}

	allUsers := make([]userDB, 0)

	for rows.Next() {
		var usr userDB
		err = rows.Scan(&usr.ID, &usr.First, &usr.Last)
		if err != nil {
			panic(err)
		}
		allUsers = append(allUsers, usr)
	}

	fmt.Println(allUsers)
	json.NewEncoder(w).Encode(allUsers)

	fmt.Println("Success!")

}

func getPeople(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		fmt.Fprintf(w, "Must be a 'GET' request")
		return
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	rows, err := db.Query("SELECT last, first from people order by last, first limit 20")
	if err != nil {
		panic(err)
	}

	allUsers := make([]userLastFirst, 0)

	for rows.Next() {
		var usr userLastFirst
		err = rows.Scan(&usr.Last, &usr.First)
		if err != nil {
			panic(err)
		}
		allUsers = append(allUsers, usr)
	}

	fmt.Println(allUsers)
	json.NewEncoder(w).Encode(allUsers)

	fmt.Println("Success!")
}

type userDB struct {
	ID    string `json:"id,omitempty"`
	First string `json:"first"`
	Last  string `json:"last"`
}

type userLastFirst struct {
	Last  string `json:"last"`
	First string `json:"first"`
}
