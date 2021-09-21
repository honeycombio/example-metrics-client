package main

import (
	"context"
)

func main() {
	ctx := context.Background()

	shutdownTraces, err := setupTraces(ctx)
	if err != nil {
		panic(err)
	}
	defer shutdownTraces()

	shutdownMetrics, err := setupMetrics(ctx)
	if err != nil {
		panic(err)
	}
	defer shutdownMetrics()

	serve()
}
