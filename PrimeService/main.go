package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func findPrime(w http.ResponseWriter, req *http.Request, inputChan chan PrimeQuery) {
	log.Println("Serving content")
	// Converting start value from URL
	number, err := strconv.Atoi(req.URL.Query().Get("val"))
	if err != nil {
		fmt.Fprintf(w, "Invalid input!\n")
		return
	}

	// Creating client channel for results
	returnChan := make(chan PrimeResponse)

	// Creating number manager
	numberManager := makeNumberManager(number)

	// Sending 5 prime queries
	for i := 0; i < 5; i++ {
		inputChan <- PrimeQuery{
			Number:  numberManager.Next().Unwrap(),
			RetChan: returnChan,
		}
	}

	// Sending a new query after each result
	foundSolution := false
	for !foundSolution {
		result := <-returnChan

		foundSolution = numberManager.CheckResult(result)

		numNext := numberManager.Next()
		if numNext.IsNone() {
			continue
		}

		inputChan <- PrimeQuery{
			Number:  numNext.Unwrap(),
			RetChan: returnChan,
		}
	}

	solution := numberManager.GetSolution()
	log.Println("Solution found")
	if solution.IsNone() {
		fmt.Fprintln(w, "No solution found!")
	} else {
		fmt.Fprintf(w, "%d\n", solution.Unwrap())
	}
}

func main() {
	inputChan, err := initServers()
	if err != nil {
		log.Fatalf("Unable to initialize server\n\tError: %s", err)
	}
	http.HandleFunc("/findPrime",
		func(w http.ResponseWriter, req *http.Request) { findPrime(w, req, inputChan) },
	)

	http.ListenAndServe(":2030", nil)
}
