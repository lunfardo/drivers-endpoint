package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

const (
	DB_HOST = "db-service"
	DB_USER = "postgres"
	DB_NAME = "drivers"
)

//this should include context.Context for distributed transactions
type context struct {
	db *sql.DB
}

func (ctx context) storeLocationDB(driverID, lat, long string, timed time.Time) error {
	// check if db is okay, if it is not, try to connect again
	err := ctx.db.Ping()
	if err != nil {
		return err
	}

	stmt, err := ctx.db.Prepare("INSERT INTO drivers_locations(driver_id,altitude,longitude,timed) VALUES($1,$2,$3,$4)")
	if err == nil {
		_, err = stmt.Exec(driverID, lat, long, timed)
	}
	return err
}

func (ctx context) handler(w http.ResponseWriter, h *http.Request) {

	err := h.ParseForm()
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "bad payload", 400)
		return
	}

	err = ctx.storeLocationDB(h.Form["driver_id"][0], h.Form["latitude"][0], h.Form["longitude"][0], time.Now())
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "internal error", 500)
		return
	}

	fmt.Printf("%v new value stored \n", time.Now())
}

func main() {
	var ctx context
	var err error
	dbinfo := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable", DB_HOST, DB_USER, DB_NAME)

	ctx.db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.db.Close()

	router := http.NewServeMux()

	//this wouldnt be standard restful, should be /api/v1/driver/{driver_id}/locations (using gorilla mux), but the payload has the driver_id
	router.HandleFunc("/api/v1/drivers/locations", ctx.handler)
	http.Handle("/", router)
	fmt.Println("Server running...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
