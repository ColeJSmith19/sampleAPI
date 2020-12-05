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
var psqlInfo string

func makeDBconnection() (*sql.DB, error) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func main() {

	var pwFlag = flag.String("pw", "", "please include a -pw flag for your database password")

	flag.Parse()

	password = *pwFlag
	if password == "" {
		fmt.Println("please include a -pw flag for your database password")
		return
	}

	psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	fmt.Println("Listening on port 18080...")
	http.HandleFunc("/", homePage)
	http.HandleFunc("/people", getPeople)
	http.HandleFunc("/person", personEndPoint)
	// http.HandleFunc("/person/{first}/{last}", getPerson)
	log.Fatal(http.ListenAndServe(":18080", nil))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}

func personEndPoint(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		getPersonQueryString(w, r)
	case "POST":
		postPerson(w, r)
	default:
		http.Error(w, "Must be a 'GET' or 'POST' request", 400)
		return
	}

}

func postPerson(w http.ResponseWriter, r *http.Request) {

	db, err := makeDBconnection()
	if err != nil {
		http.Error(w, "Something went wrong", 500)
		return
	}
	defer db.Close()

	params := r.URL.Query()

	last, lastok := params["last"]
	first, firstok := params["first"]

	if !lastok || !firstok {
		http.Error(w, "Please include 'first' and 'last' as a query parameter", 400)
		return
	}

	query := "insert into people (first, last) values ($1, $2)"
	_, err = db.Exec(query, first[0], last[0])
	if err != nil {
		http.Error(w, "Something went wrong", 500)
		return
	}

	json.NewEncoder(w).Encode(fmt.Sprintf("Successfully inserted %s %s", first[0], last[0]))

}

func getPersonQueryString(w http.ResponseWriter, r *http.Request) {

	db, err := makeDBconnection()
	if err != nil {
		http.Error(w, "Something went wrong", 500)
		return
	}
	defer db.Close()

	params := r.URL.Query()

	last, lastok := params["last"]
	first, firstok := params["first"]

	if !lastok && !firstok {
		http.Error(w, "Please include 'first' and/or 'last' as a query parameter", 400)
		return
	}

	query := ""

	if lastok && firstok {
		query = fmt.Sprintf("SELECT id, first, last from people where first = '%s' and last = '%s'", first[0], last[0])
	} else if lastok {
		query = fmt.Sprintf("SELECT id, first, last from people where last = '%s'", last[0])
	} else {
		query = fmt.Sprintf("SELECT id, first, last from people where first = '%s'", first[0])
	}

	rows, err := db.Query(query)
	if err != nil {
		fmt.Println(err)

		http.Error(w, "Something went wrong", 500)
		return
	}

	allUsers := make([]userDB, 0)

	for rows.Next() {
		var usr userDB
		err = rows.Scan(&usr.ID, &usr.First, &usr.Last)
		if err != nil {
			fmt.Println(err)

			http.Error(w, "Something went wrong", 500)
			return
		}
		allUsers = append(allUsers, usr)
	}

	if len(allUsers) == 0 {
		http.Error(w, "No data found", 404)
		return
	}

	fmt.Println(allUsers)
	json.NewEncoder(w).Encode(allUsers)

	fmt.Println("Success!")

}

// func getPerson(w http.ResponseWriter, r *http.Request) {

// db, err := makeDBconnection()
// if err != nil {
// 	http.Error(w, "Something went wrong", 500)
// 	return
// }
// defer db.Close()

// 	rows, err := db.Query("SELECT id, first, last from people order by first desc limit 20")
// 	if err != nil {
// 		panic(err)
// 	}

// 	allUsers := make([]userDB, 0)

// 	for rows.Next() {
// 		var usr userDB
// 		err = rows.Scan(&usr.ID, &usr.First, &usr.Last)
// 		if err != nil {
// 			panic(err)
// 		}
// 		allUsers = append(allUsers, usr)
// 	}

// 	fmt.Println(allUsers)
// 	json.NewEncoder(w).Encode(allUsers)

// 	fmt.Println("Success!")

// }

func getPeople(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Must be a 'GET' request", 400)
		return
	}

	// psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := makeDBconnection()
	if err != nil {
		http.Error(w, "Something went wrong", 500)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT last, first from people order by last, first")
	if err != nil {
		http.Error(w, "Something went wrong", 500)
		return
	}

	allUsers := make([]userLastFirst, 0)

	for rows.Next() {
		var usr userLastFirst
		err = rows.Scan(&usr.Last, &usr.First)
		if err != nil {
			http.Error(w, "Something went wrong", 500)
			return
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
