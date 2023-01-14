package main

import (
	"log"
)

type NumberManager struct {
	numStart        int
	primeClosest    *int
	noAnswer        []int
	resendQueue     []int
	primeCanditates map[int][2]int
}

func makeNumberManager(numStart int) NumberManager {
	log.Println("Initializing number manager")
	return NumberManager{
		numStart:        numStart,
		noAnswer:        []int{numStart},
		resendQueue:     make([]int, 0),
		primeCanditates: make(map[int][2]int),
	}
}

func (nm NumberManager) HasNext() bool {
	// A prime number has been found and all previous numbers are answered
	if nm.primeClosest != nil && *nm.primeClosest < nm.noAnswer[0] {
		return false
	}

	// 100000 numbers have been searched and no solution has been found
	if nm.noAnswer[len(nm.noAnswer)-1]-nm.numStart > 100000 {
		return false
	}

	// Continue iterating
	return true
}

func (nm *NumberManager) Next() *int {
	var numNext *int

	if len(nm.resendQueue) > 0 {
		// Take values from resend queue first
		numNext = &nm.resendQueue[0]
		nm.resendQueue = nm.resendQueue[1:]
	} else if nm.primeClosest == nil {
		// If no prime has been found continue increasing
		numOld := nm.noAnswer[len(nm.noAnswer)-1] + 1
		numNext = &numOld
		nm.noAnswer = append(nm.noAnswer, *numNext)
	}

	// Otherwise send nil for no query
	return numNext
}

func (nm *NumberManager) CheckResult(result PrimeResult) {
	log.Printf("Checking result: number=%d, result=%s\n", result.Number, result.IsPrime)

	// Resend number if error occured
	if result.Error != nil {
		nm.resendQueue = append(nm.resendQueue, result.Number)
		return
	}

	// Number is already a candidate
	count, isIn := nm.primeCanditates[result.Number]
	if isIn {
		// Increase the verification counts
		count[0] += 1
		if result.IsPrime {
			count[1] += 1
		}

		// Both verification results have been received
		if count[0] == 2 {
			delete(nm.primeCanditates, result.Number)
			if count[1] > 0 {
				if nm.primeClosest == nil {
					nm.primeClosest = &result.Number
				} else if result.Number < *nm.primeClosest {
					*nm.primeClosest = result.Number
				}
			}
			return
		}

		// Update verification counts if results are pending
		nm.primeCanditates[result.Number] = count
		return
	}

	// Find index of number in noAnswer slice and remove from slice
	idxNum := binarySearch(nm.noAnswer, result.Number)
	nm.noAnswer = append(nm.noAnswer[:idxNum], nm.noAnswer[idxNum+1:]...)

	// Resend number twice to verify the result
	if result.IsPrime {
		log.Printf("Number: %d, No Answer: %s\n", result.Number, nm.noAnswer)

		nm.primeCanditates[result.Number] = [2]int{0, 0}
		nm.resendQueue = append(nm.resendQueue, result.Number)
		nm.resendQueue = append(nm.resendQueue, result.Number)
		return

	}

}

func (nm NumberManager) GetSolution() *int {
	return nm.primeClosest
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
