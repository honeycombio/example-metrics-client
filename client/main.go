package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func randomWithinRange(min int, max int) int {
	return rand.Intn(max-min) + min
}

func generateRollCommand() string {
	return fmt.Sprintf("%dd%d", randomWithinRange(1, 100), randomWithinRange(1, 20))
}

func request() {
	start := time.Now()
	polyhedronHost, ok := os.LookupEnv("POLYHEDRON_HOST")
	if !ok {
		polyhedronHost = "polyhedron"
	}
	endpoint := fmt.Sprintf("http://%s/%s", polyhedronHost, generateRollCommand())
	resp, err := http.Get(endpoint)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("%stime: %dms\n", string(body), elapsed.Milliseconds())
}

func main() {
	concurrencyLimit := 10
	semaphoreChan := make(chan struct{}, concurrencyLimit)
	defer close(semaphoreChan)

	for {
		semaphoreChan <- struct{}{}
		go func() {
			request()
			sleepTime := time.Duration(rand.Intn(2000)) * time.Microsecond
			time.Sleep(sleepTime)
			<-semaphoreChan
		}()
	}
}
