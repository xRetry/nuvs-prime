package main

type NumberManager struct {
	numStart        int
	primeClosest    Option[int]
	noAnswer        []int
	resendQueue     []int
	primeCanditates map[int][2]int
	numLast         Option[int]
	numFirstPending int
}

func makeNumberManager(numStart int) NumberManager {
	lg.Println("[Number Manager] Initializing number manager")
	return NumberManager{
		numStart:        numStart,
		primeClosest:    None[int](),
		noAnswer:        make([]int, 0),
		resendQueue:     make([]int, 0),
		primeCanditates: make(map[int][2]int),
		numLast:         None[int](),
		numFirstPending: numStart,
	}
}

func (nm *NumberManager) Next() Option[int] {
	var numNext Option[int]

	if len(nm.resendQueue) > 0 {
		// Take values from resend queue first
		numNext = Some(nm.resendQueue[0])
		nm.resendQueue = nm.resendQueue[1:]
	} else if nm.primeClosest.IsNone() {
		// If no prime has been found continue increasing
		if nm.numLast.IsNone() {
			if nm.numStart != 2 && nm.numStart%2 == 0 {
				numNext = Some(nm.numStart + 1)
			} else {
				numNext = Some(nm.numStart)
			}
			nm.numFirstPending = numNext.Unwrap()
		} else {
			numNext = Some(nm.numLast.Unwrap() + 2)
		}
		nm.noAnswer = append(nm.noAnswer, numNext.Unwrap())
		nm.numLast = numNext
	}

	return numNext
}

func (nm *NumberManager) CheckResult(result PrimeResponse) bool {
	lg.Printf("[Number Manager] Checking result: %s\n", result)

	// Resend number if error occured
	if result.Error != nil {
		nm.resendQueue = append(nm.resendQueue, result.Number)
		return false
	}

	// Number is already a candidate
	count, isIn := nm.primeCanditates[result.Number]
	if isIn {
		// Increase the verification counts
		count[0] += 1
		if result.IsPrime {
			count[1] += 1
		}

		// Update verification counts if results are pending
		if count[0] < 2 || count[1] == 0 {
			nm.primeCanditates[result.Number] = count
			return false
		}

		// Both verification results have been received
		delete(nm.primeCanditates, result.Number)
		if count[1] > 0 {
			if nm.primeClosest.IsNone() {
				nm.primeClosest = Some(result.Number)
			} else if result.Number < nm.primeClosest.Unwrap() {
				nm.primeClosest = Some(result.Number)
			}
		}

	} else {
		if result.IsPrime {
			nm.primeCanditates[result.Number] = [2]int{0, 0}
			nm.resendQueue = append(nm.resendQueue, result.Number)
			nm.resendQueue = append(nm.resendQueue, result.Number)
			return false

		}
	}

	// Find index of number in noAnswer slice and remove from slice
	searchIdx := binarySearch(nm.noAnswer, result.Number)
	if searchIdx.IsNone() {
		return false
	}
	idxNum := searchIdx.Unwrap()
	nm.noAnswer = append(nm.noAnswer[:idxNum], nm.noAnswer[idxNum+1:]...)

	if result.Number == nm.numFirstPending {
		if len(nm.noAnswer) > 0 {
			nm.numFirstPending = nm.noAnswer[0]
		} else {
			nm.numFirstPending = nm.numStart + 1e16
		}
	}

	// A prime number has been found and all previous numbers are answered
	if nm.primeClosest.IsSome() && (len(nm.noAnswer) == 0 || nm.primeClosest.Unwrap() < nm.noAnswer[0]) {
		return true
	}

	// 100000 numbers have been searched and no solution has been found
	if nm.numLast.Unwrap()-nm.numStart > 100000 {
		return true
	}

	return false

}

func (nm NumberManager) GetSolution() Option[int] {
	return nm.primeClosest
}

func binarySearch(arr []int, key int) Option[int] {
	high := len(arr) - 1
	low := 0
	var mid int
	for low <= high {
		mid = (high + low) / 2
		if arr[mid] == key {
			return Some(mid)
		} else if arr[mid] > key {
			high = mid
		} else {
			low = mid + 1
		}
	}
	return None[int]()
}
