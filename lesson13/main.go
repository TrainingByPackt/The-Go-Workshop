package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"net/http"
)

func initialize() (*DBOps, error) {

	conn, err := sql.Open("sqlite3", "customer.db")
	if err != nil {
		return nil, fmt.Errorf("could not open db connection %v", err)
	}

	return &DBOps{Conn: conn}, nil
}

type DBOps struct {
	Conn *sql.DB
}

func (dbo *DBOps) CreateDatabase() error {

	conn := dbo.Conn
	if _, err := conn.Exec(`CREATE TABLE "CUSTOMER" (
	"FIRSTNAME"	TEXT,
	"ID"	INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
	"ADDRESS"	TEXT,
	"LASTNAME"	TEXT,
	"STATE"	TEXT
    )`); err != nil {
		return fmt.Errorf("error occured when attempting to create customer table %v", err)
	}

	return nil

}

func (dbo *DBOps) getRecords() ([]ListRecordRequest, error) {

	var resList []ListRecordRequest
	//
	var (
		firstName string
		id        int
		lastName  string
		address   string
		state     string
	)

	conn := dbo.Conn

	row, _ := conn.Query("SELECT * FROM CUSTOMER")

	for row.Next() {
		if err := row.Scan(&firstName, &id, &address, &lastName, &state); err != nil {
			return nil, fmt.Errorf("erro scanning result set %v", err)
		}

		if firstName != "" || lastName != "" || address != "" || id != 0 {
			rr := ListRecordRequest{
				ID:        id,
				FirstName: firstName,
				Address:   address,
				LastName:  lastName,
				State:     state,
			}

			resList = append(resList, rr)
		}

	}

	return resList, nil

}

func (dbo *DBOps) createRecord(data *CreateRecordRequest) (*CreateRecordResponse, error) {

	conn := dbo.Conn

	statement, err := conn.Prepare("INSERT INTO CUSTOMER (FIRSTNAME, LASTNAME, ADDRESS, STATE) VALUES (?,?,?,?)")
	if err != nil {
		return &CreateRecordResponse{Status: err.Error()}, fmt.Errorf("could not prepare statement %v", err)
	}

	if _, err := statement.Exec(data.FirstName, data.LastName, data.City, data.State); err != nil {
		return &CreateRecordResponse{Status: err.Error()}, fmt.Errorf("statement execution failed %v", err)
	}

	return &CreateRecordResponse{Status: "record created"}, nil

}

type CreateRecordRequest struct {
	FirstName string `json:"first"`
	LastName  string `json:"last"`
	Address   string `json:"address"`
	City      string `json:"city,omitempty"`
	State     string `json:"state"`
}

type ListRecordRequest struct {
	ID        int
	FirstName string
	LastName  string
	Address   string
	City      string
	State     string
}

type CreateRecordResponse struct {
	Status string
}

func createCustomerHandler(w http.ResponseWriter, r *http.Request) {


	var request CreateRecordRequest
	db, err := initialize()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	f, err := ioutil.ReadAll(r.Body)
	fmt.Println(string(f))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	switch r.Header.Get("Content-Type") {
	case "application/json":
		if err := json.Unmarshal(f, &request); err != nil {
			fmt.Printf("%+v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}



		resp, err := db.createRecord(&request)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		respJson, err := json.Marshal(resp)
		if err != nil{
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Write(respJson)
	}
}

func getCustomerHandler(w http.ResponseWriter, r *http.Request) {

	db, err := initialize()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	records, err := db.getRecords()

	switch r.Header.Get("Content-Type") {
	case "application/json":
		for _, recs := range records {
			resp, err := json.Marshal(recs)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			w.Write(resp)

		}
	case "text/html":

	}

}

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte("welcome to the booking agency")); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	r.HandleFunc("/createcustomer", createCustomerHandler)
	r.HandleFunc("/getcustomers", getCustomerHandler)

	srv := http.Server{
		Addr:    ":3000",
		Handler: r,
	}

	log.Fatal(srv.ListenAndServe())

}
