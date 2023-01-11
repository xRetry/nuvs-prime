package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type PrimeResult struct {
	Number  int
	IsPrime bool
	Error   error
}

type PrimeQuery struct {
	Number  int
	RetChan chan PrimeResult
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
		go makeServerConnection(addr, inputChan)
	}

	return inputChan, nil
}

func findAvailableServers() ([]string, error) {
	log.Println("Finding available services")
	client := http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get("http://10.21.0.13:2020/api/v1.0/active-http-services")
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	var availServices []AvailableService
	err = json.NewDecoder(resp.Body).Decode(&availServices)
	if err != nil {
		return []string{}, err
	}

	addrs := make([]string, 0)
	for _, service := range availServices {
		addrs = append(addrs, service.Ip)
	}
	log.Printf("Services found: %d\n", len(addrs))
	return addrs, nil
}

func makeServerConnection(addr string, inputChan <-chan PrimeQuery) {
	log.Printf("Connection started on address: %s\n", addr)
	for {
		query := <-inputChan
		resp, err := http.Get(fmt.Sprintf("http://%s:2000/isPrime?val=%d", addr, query.Number))

		if err != nil {
			query.RetChan <- PrimeResult{
				Number:  query.Number,
				IsPrime: false,
				Error:   err,
			}
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			query.RetChan <- PrimeResult{
				Number:  query.Number,
				IsPrime: false,
				Error:   err,
			}
			continue
		}

		if string(body) == "TRUE" {
			query.RetChan <- PrimeResult{
				Number:  query.Number,
				IsPrime: true,
				Error:   nil,
			}
			continue
		}

		query.RetChan <- PrimeResult{
			Number:  query.Number,
			IsPrime: false,
			Error:   nil,
		}
	}
}
