package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func findPrime(w http.ResponseWriter, req *http.Request, inputChan chan PrimeQuery) {
	log.Println("Serving content")
	number, err := strconv.Atoi(req.URL.Query().Get("val"))
	if err != nil {
		fmt.Fprintf(w, "Invalid input!\n")
		return
	}

	returnChan := make(chan PrimeResult)

	numberManager := makeNumberManager(number)

	for i := 0; i < 5; i++ {
		inputChan <- PrimeQuery{
			Number:  numberManager.Next(),
			RetChan: returnChan,
		}
	}

	for numberManager.HasNext() {
		result := <-returnChan

		ptrClosest := numberManager.CheckResult(result)
		if ptrClosest == nil {
			inputChan <- PrimeQuery{
				Number:  numberManager.Next(),
				RetChan: returnChan,
			}
			continue
		}

		numClosest := *ptrClosest

		// TODO: Verify result
		// TODO: Interrupt other queries

		fmt.Fprintf(w, "%d\n", numClosest)
		break
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