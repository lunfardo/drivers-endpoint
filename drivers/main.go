package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	DEFAULT_DRIVERS_COUNT = "5"
	DEFAULT_TIME_MS       = "1200"
	DEFAULT_API_HOSTNAME  = "api-service"
)

func getenv(key, dft string) string {
	val := os.Getenv(key)
	if len(val) > 0 {
		return val
	}
	return dft
}

func driverPositionPoster(driverID int) {

	driveIDstr := strconv.Itoa(driverID)
	ApiHost := getenv("DRIVER_API_HOST", DEFAULT_API_HOSTNAME)
	ApiEndpoint := "http://" + ApiHost + ":8080/api/v1/drivers/locations"
	httpClientTimeoutMs, err := strconv.Atoi(getenv("DRIVER_HTTP_TIMEOUT", DEFAULT_TIME_MS))
	if err != nil {
		log.Fatal(err)
	}
	//is always good practices having a http client with timeout, it had to be added in go 1.3
	client := &http.Client{
		Timeout: time.Millisecond * time.Duration(httpClientTimeoutMs),
	}

	for true {
		//values are passed in a post-form-like way
		resp, err := client.PostForm(ApiEndpoint,
			url.Values{
				"driver_id": {driveIDstr},
				"latitude":  {"123"},
				"longitude": {"123"},
			})
		//timeout and bad database connections should trigger first part of the conditional
		if err != nil {
			log.Printf("Driver %v posted position and FAILED with: %v", driverID, err.Error())
		} else {
			fmt.Printf("Driver %v posted position and recieved HTTP status code: %v \n", driverID, resp.StatusCode)
		}
		time.Sleep(time.Second * 5)
	}
}

func main() {
	var total_drivers int
	var wg sync.WaitGroup
	var err error

	total_drivers, err = strconv.Atoi(getenv("GO_TOTAL_DRIVERS", DEFAULT_DRIVERS_COUNT))

	if err != nil {
		log.Fatal(err)
	}

	wg.Add(total_drivers)
	for i := 1; i <= total_drivers; i++ {
		go driverPositionPoster(i)
	}
	//when parent die, children die too, and we dont want that. Anyways, this process will never end of waiting since the "for true" in the children process
	wg.Wait()

}
