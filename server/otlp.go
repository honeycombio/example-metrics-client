package main

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	metricglobal "go.opentelemetry.io/otel/metric/global"
	metriccontroller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"google.golang.org/grpc/credentials"
)

func createOTLPExporter(ctx context.Context, datasetName string) (*otlp.Exporter, error) {
	otlp_endpoint := os.Getenv("OTLP_ENDPOINT")
	if otlp_endpoint == "" {
		otlp_endpoint = "api.honeycomb.io:443"
	}

	return otlp.NewExporter(ctx, otlpgrpc.NewDriver(
		otlpgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
		otlpgrpc.WithEndpoint(otlp_endpoint), // otel-collector running as agent on this host (4317 is the default grpc port)
		otlpgrpc.WithHeaders(map[string]string{
			"x-honeycomb-team":    os.Getenv("HONEYCOMB_API_KEY"),
			"x-honeycomb-dataset": datasetName,
		}),
	))
}

func setupTraces(ctx context.Context) (func(), error) {
	tracesDatasetName := os.Getenv("HONEYCOMB_TRACES_DATASET")
	if tracesDatasetName == "" {
		tracesDatasetName = "polyhedron_traces"
	}

	exporter, err := createOTLPExporter(ctx, tracesDatasetName)
	if err != nil {
		return nil, err
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.5)),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tp)
	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			panic(err)
		}
	}, nil
}

func setupMetrics(ctx context.Context) (func(), error) {
	metricsDatasetName := os.Getenv("HONEYCOMB_METRICS_DATASET")
	if metricsDatasetName == "" {
		metricsDatasetName = "polyhedron_metrics"
	}

	exporter, err := createOTLPExporter(ctx, metricsDatasetName)
	if err != nil {
		return nil, err
	}

	mc := metriccontroller.New(
		processor.New(simple.NewWithHistogramDistribution(), exporter),
		metriccontroller.WithCollectPeriod(10*time.Second),
		metriccontroller.WithExporter(exporter),
		metriccontroller.WithResource(resource.NewWithAttributes(semconv.ServiceNameKey.String("polyhedron"))),
	)
	metricglobal.SetMeterProvider(mc.MeterProvider())

	// Capture runtime metrics
	if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second)); err != nil {
		panic(err)
	}

	// Handle this error in a sensible manner where possible
	return func() {
		if err := mc.Stop(ctx); err != nil {
			panic(err)
		}
	}, mc.Start(ctx)
}
