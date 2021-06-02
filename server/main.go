package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/justinian/dice"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	metricglobal "go.opentelemetry.io/otel/metric/global"
	metriccontroller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

func handleRoll(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	span := trace.SpanFromContext(ctx)

	nameKey := attribute.Key("name")
	dieRequestKey := attribute.Key("dieRoll.request")
	dieResultKey := attribute.Key("dieRoll.result")
	errorKey := attribute.Key("error")

	ctx = baggage.ContextWithValues(ctx, nameKey.String("roll"))

	span.SetAttributes(nameKey.String("roll"), dieRequestKey.String(req.URL.Path))
	w.Header().Add("Content-Type", "text/html")

	if req.URL.Path == "/" {
		fmt.Fprintf(w, "<h1>Welcome to polyhedron!</h1><p>Try <a href=\"/1d6\">/1d6</a></p>")
		return
	}

	result, reason, err := dice.Roll(req.URL.Path)
	if err != nil {
		fmt.Fprintf(w, "Error: %v\n", err)
		span.SetAttributes(errorKey.String(err.Error()))
		return
	}

	span.SetAttributes(dieResultKey.Int(result.Int()))
	fmt.Fprintf(w, "<b>Roll:</b> %s<br /><b>Result:</b> %d<br />%s", result.Description(), result.Int(), reason)
}

func main() {
	ctx := context.Background()

	exporter, err := createMetricsExporter(ctx)
	if err != nil {
		panic(err)
	}

	shutdownTraces, err := setupTraces(ctx, exporter)
	if err != nil {
		panic(err)
	}
	defer shutdownTraces()

	shutdownMetrics, err := setupMetrics(ctx, exporter)
	if err != nil {
		panic(err)
	}
	defer shutdownMetrics()

	http.Handle("/", otelhttp.NewHandler(http.HandlerFunc(handleRoll), "Roll"))
	if err = http.ListenAndServe(":8090", nil); err != nil {
		panic(err)
	}
}

var webserverResource = resource.NewWithAttributes(semconv.ServiceNameKey.String("webserver"))

func createMetricsExporter(ctx context.Context) (*otlp.Exporter, error) {
	return otlp.NewExporter(ctx, otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),                 // insecure because sending to localhost
		otlpgrpc.WithEndpoint("localhost:4317"), // otel-collector running as agent on this host (4317 is the default grpc port)
		otlpgrpc.WithHeaders(map[string]string{"ContentType": "application/grpc"}),
	))
}

func setupTraces(ctx context.Context, exporter *otlp.Exporter) (func(), error) {
	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(webserverResource),
	)
	otel.SetTracerProvider(tp)
	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			panic(err)
		}
	}, nil
}

func setupMetrics(ctx context.Context, exporter *otlp.Exporter) (func(), error) {
	mc := metriccontroller.New(
		processor.New(simple.NewWithExactDistribution(), exporter),
		metriccontroller.WithExporter(exporter),
		metriccontroller.WithCollectPeriod(5*time.Second),
		metriccontroller.WithResource(webserverResource),
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
