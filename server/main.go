package main

import (
	"context"
)

func main() {
	ctx := context.Background()

	exporter, err := createOTLPExporter(ctx)
	if err != nil {
		panic(err)
	}

	// shutdownTraces, err := setupTraces(ctx, exporter)
	// if err != nil {
	// 	panic(err)
	// }
	// defer shutdownTraces()

	shutdownMetrics, err := setupMetrics(ctx, exporter)
	if err != nil {
		panic(err)
	}
	defer shutdownMetrics()

	serve()
}
