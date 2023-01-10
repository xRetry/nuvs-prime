package main

import (
    "strconv"
    "net/http"
    "fmt"
    "io/ioutil"
    "strings"
	"math"
)

type ReturnChannel chan int, bool, error
type InputChannel chan ReturnChannel, int


func initServers() InputChannel {
	availAddrs, err := findAvailableServers()
	if err != nil {
		fmt.Fprintf(w, "Address service not reachable\n")
		return
	}

	inputChan := make(InputChannel)

	for _, addr := range availAddrs {
		go spawnServer(addr, inputChan)		
	}

	return inputChan
}

type NumberGenerator struct {
	num_start int
	num_closest *int
	num_last *int
	no_answer map[int]Struct{}
}

func makeNumberGen(num_start int) NumberGenerator {
	return NumberGenerator{
		num_start: num_start,
		num_closest: nil,
		num_last: nil
		no_answer: make(map[int]Struct{})
	}
}

func (g *NumberGenerator) Next() int {
	num_new := g.num_start+1
	switch last := g.num_last; {
		case last == nil:
			num_new = g.num_start
		case *last < g.num_start:
			num_new = g.num_last + 2*math.Abs(g.num_last - g.num_start) + 1
		case *last > g.num_start
			num_new = g.num_last - 2*math.Abs(g.num_last - g.num_start)
 	}

	*g.num_last = num_new
	g.no_answer[num_last] = Struct{}
	return num_last
}

func (g *NumberGenerator) CheckResult(number int, isPrime bool) bool {
	if
	g.no_an
}

func findPrime(w http.ResponseWriter, req *http.Request, inputChan InputChannel) {
	number, err := strconv.Atoi(req.URL.Query().Get("val"))
	if err != nil {
		fmt.Fprintf(w, "Invalid input!\n")
		return
	}

    fmt.Println(number)

	returnChan := make(ReturnChannel)
	
	numberGen := makeNumberGen(number)

	for i:=0; i<5; i++ {
		inputChan <- returnChan, numberGen.Next()
	}
    
	for iterNumbers.HasNext() {
		number, isPrime, err <- returnChan

		isClosest := numberGen.CheckResult(number, isPrime) // TODO: Add method
		if isClosest {
			fmt.Fprintf(w, "%d\n", number)
			break
		}

		inputChan <- returnChan, iterNumbers.next(1)
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

func spawnServer(addr string, inputChan <-InputChannel) {
	for {
		returnChan, number := <- inputChan
        resp, err := http.Get(fmt.Sprintf("http://%s:2000/isPrime?val=%d", addr, number))

        if err != nil {
            returnChan <- number, false, fmt.Errorf("Unable to connect to server")
			continue
        }

        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            returnChan <- number, false, fmt.Errorf("Unable to read body")
			continue
        }

        if string(body) == "true" {
            returnChan <- number, true, nil
			continue
        }

		returnChan <- number, false, nil
	}
}


func main() {
	inputChan = initServers()
    http.HandleFunc("/findPrime", 
		func(w http.ResponseWriter, req *http.Request) { findPrime(w, req, inputChan) } 
	)
}
