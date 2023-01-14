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
	returnChan := make(chan PrimeResult)

	// Creating number manager
	numberManager := makeNumberManager(number)

	// Sending 5 prime queries
	for i := 0; i < 5; i++ {
		inputChan <- PrimeQuery{
			Number:  *numberManager.Next(),
			RetChan: returnChan,
		}
	}

	// Sending a new query after each result
	for numberManager.HasNext() {
		result := <-returnChan

		numNext := numberManager.Next()

		if numNext == nil {
			continue
		}

		numberManager.CheckResult(result)
		inputChan <- PrimeQuery{
			Number:  *numNext,
			RetChan: returnChan,
		}
	}

	solution := numberManager.GetSolution()
	if solution == nil {
		fmt.Fprintln(w, "No solution found!")
	} else {
		fmt.Fprintf(w, "%d\n", *solution)
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
