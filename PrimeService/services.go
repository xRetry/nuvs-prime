package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type PrimeResponse struct {
	Number  int
	IsPrime Result[bool]
}

type PrimeQuery struct {
	Number  int
	RetChan chan PrimeResponse
}

type AvailableService struct {
	Ip     string `json:"ip"`
	Header string `json:"header"`
	Valid  int    `json:"valid"`
}

func initServers() (chan PrimeQuery, error) {
	log.Println("Starting service initialization")
	availAddrs, err := findAvailableServers()
	if err != nil {
		return nil, err
	}

	inputChan := make(chan PrimeQuery)

	for _, addr := range availAddrs {
		go makeServerHandler(addr, inputChan)
	}

	return inputChan, nil
}

func findAvailableServers() ([]string, error) {
	log.Println("Finding available services")
	client := http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get("http://10.21.0.13:2020/api/v1.0/active-http-services")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var availServices []AvailableService
	err = json.NewDecoder(resp.Body).Decode(&availServices)
	if err != nil {
		return nil, err
	}

	addrs := make([]string, 0)
	for _, service := range availServices {
		addrs = append(addrs, service.Ip)
	}
	log.Printf("Services found: %d\n", len(addrs))
	return addrs, nil
}

func makeServerHandler(addr string, inputChan <-chan PrimeQuery) {
	log.Printf("Connection started on address: %s\n", addr)

	for {
		log.Printf("%s waiting..\n", addr)
		query := <-inputChan

		log.Printf("%s got query: %d\n", addr, query.Number)
		go queryServer(addr, query)
	}
}

func queryServer(addr string, query PrimeQuery) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s:2000/isPrime?val=%d", addr, query.Number))
	log.Printf("%s got response for %d\n", addr, query.Number)
	if err != nil {
		log.Printf("Connection error %s\n", err)
		query.RetChan <- PrimeResponse{
			Number:  query.Number,
			IsPrime: Err[bool](err),
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if body != nil {
		query.RetChan <- PrimeResponse{
			Number:  query.Number,
			IsPrime: Err[bool](err),
		}
	}

	if strings.ToLower(string(body)) == "true" {
		query.RetChan <- PrimeResponse{
			Number:  query.Number,
			IsPrime: Ok(true),
		}
	}

	query.RetChan <- PrimeResponse{
		Number:  query.Number,
		IsPrime: Ok(false),
	}

}
