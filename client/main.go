package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

func request() {
	resp, err := http.Get("http://polyhedron/2d6")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	// fmt.Printf("Result: %s\n", string(body))
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
