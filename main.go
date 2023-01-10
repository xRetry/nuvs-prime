package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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

func initServers() (chan PrimeQuery, error) {
	availAddrs, err := findAvailableServers()
	if err != nil {
		return nil, err
	}

	inputChan := make(chan PrimeQuery)

	for _, addr := range *availAddrs {
		go makeServerConnection(addr, inputChan)
	}

	return inputChan, nil
}

type NumberGenerator struct {
	numStart   int
	numClosest *int
	noAnswer   []int
}

func makeNumberGen(numStart int) NumberGenerator {
	return NumberGenerator{
		numStart:   numStart,
		numClosest: nil,
		noAnswer:   []int{numStart - 1},
	}
}

func (g *NumberGenerator) Next() int {
	numNext := g.noAnswer[len(g.noAnswer)-1] + 1
	g.noAnswer = append(g.noAnswer, numNext)
	return numNext
}

// TODO: Change to return prime number instead of bool
func (g *NumberGenerator) CheckResult(result PrimeResult) bool {
	if result.IsPrime {
		if result.Number < *g.numClosest {
			*g.numClosest = result.Number
		}
	}

	idxNum := binarySearch(g.noAnswer, result.Number)
	if idxNum == 0 {
		return true
	}

	g.noAnswer = append(g.noAnswer[:idxNum], g.noAnswer[idxNum+1:]...)
	return false
}

func binarySearch(arr []int, key int) int {
	high := len(arr) - 1
	low := 0
	var mid int
	for low <= high {
		mid = (high + low) / 2
		if arr[mid] == key {
			return mid
		} else if arr[mid] > key {
			high = mid
		} else {
			low = mid + 1
		}
	}
	return -1
}

func findPrime(w http.ResponseWriter, req *http.Request, inputChan chan PrimeQuery) {
	number, err := strconv.Atoi(req.URL.Query().Get("val"))
	if err != nil {
		fmt.Fprintf(w, "Invalid input!\n")
		return
	}

	fmt.Println(number)

	returnChan := make(chan PrimeResult)

	numberGen := makeNumberGen(number)

	for i := 0; i < 5; i++ {
		inputChan <- PrimeQuery{
			Number:  numberGen.Next(),
			RetChan: returnChan,
		}
	}

	for numberGen.HasNext() {
		result := <-returnChan

		isClosest := numberGen.CheckResult(result) // TODO: Add method
		if isClosest {
			fmt.Fprintf(w, "%d\n", number)
			break
		}

		inputChan <- PrimeQuery{
			Number:  numberGen.Next(),
			RetChan: returnChan,
		}
	}

}

func findAvailableServers() (*[]string, error) {

	resp, err := http.Get("http://10.21.0.13:2020/api/v1.0/active-http-services")
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to server")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read body")
	}

	//TODO: Parse json, create IP list
	return strings.Split(string(body), "\n")[0], nil
}

func makeServerConnection(addr string, inputChan <-chan PrimeQuery) {
	for {
		query := <-inputChan
		resp, err := http.Get(fmt.Sprintf("http://%s:2000/isPrime?val=%d", addr, query.Number))

		if err != nil {
			query.RetChan <- PrimeResult{
				Number:  query.Number,
				IsPrime: false,
				Error:   fmt.Errorf("Unable to connect to server"),
			}
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			query.RetChan <- PrimeResult{
				Number:  query.Number,
				IsPrime: false,
				Error:   fmt.Errorf("Unable to read body"),
			}
			continue
		}

		if string(body) == "true" {
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

func main() {
	inputChan, err := initServers()
	if err == nil {
		// TODO: Handle error
	}
	http.HandleFunc("/findPrime",
		func(w http.ResponseWriter, req *http.Request) { findPrime(w, req, inputChan) },
	)
}
