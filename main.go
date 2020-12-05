package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

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
	http.HandleFunc("/person/", getOrPostPerson)
	// http.HandleFunc("/queryperson", queryPersonEndPoint)
	log.Fatal(http.ListenAndServe(":18080", nil))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}

func getOrPostPerson(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/person/")
	params := strings.Split(path, "/")

	if len(params) != 2 {
		http.Error(w, "Requests must include a last and a first name\n/person/:last/:first", 400)
		return
	}

	lastName := params[0]
	firstName := params[1]

	switch r.Method {
	case "GET":
		getPerson(w, r, lastName, firstName)
	case "POST":
		postPerson(w, r, lastName, firstName)
	default:
		http.Error(w, "Must be a 'GET' or 'POST' request", 400)
		return
	}

}

func getPerson(w http.ResponseWriter, r *http.Request, last, first string) {

	db, err := makeDBconnection()
	if err != nil {
		http.Error(w, "Something went wrong", 500)
		return
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT id, last, first from people where first = '%s' and last = '%s'", first, last)

	rows, err := db.Query(query)
	if err != nil {
		fmt.Println(err)

		http.Error(w, "Something went wrong", 500)
		return
	}

	allUsers := make([]userDB, 0)

	for rows.Next() {
		var usr userDB
		err = rows.Scan(&usr.ID, &usr.Last, &usr.First)
		if err != nil {
			fmt.Println(err)

			http.Error(w, "Something went wrong", 500)
			return
		}
		allUsers = append(allUsers, usr)
	}

	if len(allUsers) == 0 {
		http.Error(w, fmt.Sprintf("/person/%s/%s does not exist", last, first), 404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allUsers)

}

func postPerson(w http.ResponseWriter, r *http.Request, last, first string) {

	db, err := makeDBconnection()
	if err != nil {
		http.Error(w, "Something went wrong", 500)
		return
	}
	defer db.Close()

	query := "insert into people (first, last) values ($1, $2)"
	_, err = db.Exec(query, first, last)
	if err != nil {
		http.Error(w, "Something went wrong", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fmt.Sprintf("Successfully inserted %s %s", first, last))

}

func getPeople(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Must be a 'GET' request", 400)
		return
	}

	db, err := makeDBconnection()
	if err != nil {
		http.Error(w, "Something went wrong", 500)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT last || ', ' || first as name from people order by last, first")
	if err != nil {
		http.Error(w, "Something went wrong", 500)
		return
	}

	allNames := make([]string, 0)

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			http.Error(w, "Something went wrong", 500)
			return
		}
		allNames = append(allNames, name)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allNames)

}

type userDB struct {
	ID    string `json:"id,omitempty"`
	Last  string `json:"last"`
	First string `json:"first"`
}

type userLastFirst struct {
	Last  string `json:"last"`
	First string `json:"first"`
}

// func queryPersonEndPoint(w http.ResponseWriter, r *http.Request) {

// 	switch r.Method {
// 	case "GET":
// 		getPersonQueryString(w, r)
// 	case "POST":
// 		postPersonQueryString(w, r)
// 	default:
// 		http.Error(w, "Must be a 'GET' or 'POST' request", 400)
// 		return
// 	}

// }

// func postPersonQueryString(w http.ResponseWriter, r *http.Request) {

// 	db, err := makeDBconnection()
// 	if err != nil {
// 		http.Error(w, "Something went wrong", 500)
// 		return
// 	}
// 	defer db.Close()

// 	params := r.URL.Query()

// 	last, lastok := params["last"]
// 	first, firstok := params["first"]

// 	if !lastok || !firstok {
// 		http.Error(w, "Please include 'first' and 'last' as a query parameter", 400)
// 		return
// 	}

// 	query := "insert into people (first, last) values ($1, $2)"
// 	_, err = db.Exec(query, first[0], last[0])
// 	if err != nil {
// 		http.Error(w, "Something went wrong", 500)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(fmt.Sprintf("Successfully inserted %s %s", first[0], last[0]))

// }

// func getPersonQueryString(w http.ResponseWriter, r *http.Request) {

// 	db, err := makeDBconnection()
// 	if err != nil {
// 		http.Error(w, "Something went wrong", 500)
// 		return
// 	}
// 	defer db.Close()

// 	params := r.URL.Query()

// 	last, lastok := params["last"]
// 	first, firstok := params["first"]

// 	if !lastok && !firstok {
// 		http.Error(w, "Please include 'first' and/or 'last' as a query parameter", 400)
// 		return
// 	}

// 	query := ""

// 	if lastok && firstok {
// 		query = fmt.Sprintf("SELECT id, first, last from people where first = '%s' and last = '%s'", first[0], last[0])
// 	} else if lastok {
// 		query = fmt.Sprintf("SELECT id, first, last from people where last = '%s'", last[0])
// 	} else {
// 		query = fmt.Sprintf("SELECT id, first, last from people where first = '%s'", first[0])
// 	}

// 	rows, err := db.Query(query)
// 	if err != nil {
// 		fmt.Println(err)

// 		http.Error(w, "Something went wrong", 500)
// 		return
// 	}

// 	allUsers := make([]userDB, 0)

// 	for rows.Next() {
// 		var usr userDB
// 		err = rows.Scan(&usr.ID, &usr.Last, &usr.First)
// 		if err != nil {
// 			fmt.Println(err)

// 			http.Error(w, "Something went wrong", 500)
// 			return
// 		}
// 		allUsers = append(allUsers, usr)
// 	}

// 	if len(allUsers) == 0 {
// 		http.Error(w, "No data found", 404)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(allUsers)
// }
