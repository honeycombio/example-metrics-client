package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

func randomWithinRange(min int, max int) int {
	return rand.Intn(max-min) + min
}

func generateRollCommand() string {
	return fmt.Sprintf("%dd%d", randomWithinRange(1, 100), randomWithinRange(1, 20))
}

func request() {
	endpoint := fmt.Sprintf("http://polyhedron/%s", generateRollCommand())
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
	fmt.Printf("Result: %s\n", string(body))
}

func main() {
	concurrencyLimit := 15
	semaphoreChan := make(chan struct{}, concurrencyLimit)
	defer close(semaphoreChan)

	for {
		semaphoreChan <- struct{}{}
		go func() {
			request()
			sleepTime := time.Duration(rand.Intn(1000)) * time.Millisecond
			time.Sleep(sleepTime)
			<-semaphoreChan
		}()
	}
}
