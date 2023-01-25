package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type PrimeResponse struct {
	Number  int
	IsPrime bool
	Error   error
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

type ServiceManager struct {
	activeServices map[string]chan bool
	inputChan      chan PrimeQuery
}

func makeServiceManger() *ServiceManager {
	return &ServiceManager{
		activeServices: make(map[string]chan bool),
		inputChan:      make(chan PrimeQuery),
	}
}

type NoServiceError struct{}

func (e NoServiceError) Error() string {
	return "No prime service available"
}

func (sm *ServiceManager) updateServices() error {
	lg.Println("[Service Manager] Starting service initialization")
	availAddrs, err := findAvailableServers()
	if err != nil {
		return err
	}

	newActiveServices := make(map[string]chan bool)

	for _, addr := range availAddrs {
		isOk := verifyService(addr)
		quitChan, isActive := sm.activeServices[addr]

		if !isOk {
			continue
		}

		if isActive {
			delete(sm.activeServices, addr)
		}

		newActiveServices[addr] = quitChan
		go handleServiceConnection(addr, sm.inputChan, quitChan)
	}

	//Close invalid services
	for _, quit := range sm.activeServices {
		quit <- true
	}

	if len(newActiveServices) == 0 {
		quitChan := make(chan bool)
		newActiveServices["Inactive"] = quitChan
		go handleNoServiceActive(sm.inputChan, quitChan)
	}

	sm.activeServices = newActiveServices

	return nil
}

func verifyService(addr string) bool {
	testNumbers := map[int]bool{
		11: true,
		12: false,
		21: false,
		23: true,
	}
	retChan := make(chan PrimeResponse, 1)

	testQueries := make([]PrimeQuery, 0)
	for num := range testNumbers {
		testQueries = append(
			testQueries,
			PrimeQuery{
				Number:  num,
				RetChan: retChan,
			},
		)
	}

	for _, query := range testQueries {
		queryServer(addr, query)
		resp := <-retChan

		if resp.Error != nil || resp.IsPrime != testNumbers[resp.Number] {
			return false
		}
	}

	return true
}

func findAvailableServers() ([]string, error) {
	lg.Println("[Service Manager] Finding available services")
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
	lg.Printf("[Service Manager] Services found: %d\n", len(addrs))
	return addrs, nil
}

func handleNoServiceActive(inputChan <-chan PrimeQuery, quitChan <-chan bool) {
	lg.Println("[Service Manager] No service active handler started")
	for {
		query := <-inputChan
		query.RetChan <- PrimeResponse{
			Number:  query.Number,
			IsPrime: false,
			Error:   NoServiceError{},
		}

	}
}

func handleServiceConnection(addr string, inputChan <-chan PrimeQuery, quitChan <-chan bool) {
	lg.Printf("[Service Manager] Connection started on address: %s\n", addr)

	for {
		lg.Printf("[Service Manager] %s waiting..\n", addr)
		select {
		case <-quitChan:
			break
		case query := <-inputChan:
			lg.Printf("[Service Manager] %s got query: %d\n", addr, query.Number)

			go queryServer(addr, query)
		}
	}
}

func queryServer(addr string, query PrimeQuery) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s:2000/isPrime?val=%d", addr, query.Number))
	lg.Printf("[Service Manager] %s got response for %d\n", addr, query.Number)
	if err != nil {
		lg.Printf("[Service Manager] Connection error %s\n", err)
		query.RetChan <- PrimeResponse{
			Number:  query.Number,
			IsPrime: false,
			Error:   err,
		}
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		query.RetChan <- PrimeResponse{
			Number:  query.Number,
			IsPrime: false,
			Error:   err,
		}
		return
	}

	if strings.ToLower(string(body)) == "true\n" {
		query.RetChan <- PrimeResponse{
			Number:  query.Number,
			IsPrime: true,
			Error:   nil,
		}
		return
	}

	if strings.ToLower(string(body)) == "false\n" {
		query.RetChan <- PrimeResponse{
			Number:  query.Number,
			IsPrime: false,
			Error:   nil,
		}
		return
	}

	query.RetChan <- PrimeResponse{
		Number:  query.Number,
		IsPrime: false,
		Error:   errors.New("Response body not true or false"),
	}

}
