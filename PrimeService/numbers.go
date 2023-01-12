package main

import (
	"log"
)

type NumberManager struct {
	numStart     int
	primeClosest *int
	noAnswer     []int
}

func makeNumberManager(numStart int) NumberManager {
	log.Println("Initializing number manager")
	return NumberManager{
		numStart:     numStart,
		primeClosest: nil,
		noAnswer:     []int{numStart - 1},
	}
}

func (g NumberManager) HasNext() bool {
	if g.noAnswer[len(g.noAnswer)-1]-g.numStart > 100000 {
		return false
	}

	return true
}

func (g *NumberManager) Next() int {
	numNext := g.noAnswer[len(g.noAnswer)-1] + 1
	g.noAnswer = append(g.noAnswer, numNext)
	return numNext
}

func (g *NumberManager) CheckResult(result PrimeResult) *int {
	log.Printf("Checking result: number=%d, result=%s\n", result.Number, result.IsPrime)

	idxNum := binarySearch(g.noAnswer, result.Number)
	g.noAnswer = append(g.noAnswer[:idxNum], g.noAnswer[idxNum+1:]...)

	if result.IsPrime {
		log.Printf("Number: %d, No Answer: %s\n", result.IsPrime, g.noAnswer)
		if g.primeClosest == nil {
			g.primeClosest = &result.Number
		} else if result.Number < *g.primeClosest {
			*g.primeClosest = result.Number
		}

		if *g.primeClosest < g.noAnswer[0] {
			return g.primeClosest
		}
	}

	return nil
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
