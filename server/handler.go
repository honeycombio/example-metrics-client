package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/justinian/dice"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("polyhedron")
var errorKey = attribute.Key("error")

var requestList []string

var rollResultRecorder metric.Int64ValueRecorder
var rollQtyRecorder metric.Int64Counter

func init() {
	meter := global.Meter("httpHandler")
	rollQtyRecorder = metric.Must(meter).NewInt64Counter("requests_received", metric.WithDescription("Measures number of requests received"))

	var err error
	rollResultRecorder, err = meter.NewInt64ValueRecorder("rollResult")
	if err != nil {
		panic(err)
	}
}

func serve() {
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		span := trace.SpanFromContext(ctx)

		w.Header().Add("Content-Type", "text/html")

		if req.URL.Path == "/" {
			fmt.Fprint(w, "<h1>Welcome to polyhedron!</h1><p>Try <a href=\"/1d6\">/1d6</a></p>\n")
			return
		}

		err, response := handleDiceRoll(ctx, req)
		if err != nil {
			span.SetAttributes(errorKey.String(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err.Error())
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s\n", response)
		}
	}))
	if err := http.ListenAndServe(":8092", nil); err != nil {
		panic(err)
	}
}

var rollMatcher = regexp.MustCompile(`\d+d\d+`)
var rollRequestKey = attribute.Key("diceRoll.request")
var rollResultValueKey = attribute.Key("diceRoll.result_value")
var rollResultReasonKey = attribute.Key("diceRoll.result_reason")

func handleDiceRoll(ctx context.Context, req *http.Request) (error, string) {
	ctx, span := tracer.Start(ctx, "handleDiceRoll")
	defer span.End()

	rollRequest := rollMatcher.FindString(req.URL.Path)
	span.SetAttributes(rollRequestKey.String(rollRequest))
	if rollRequest == "" {
		return fmt.Errorf("no roll provided"), ""
	}

	result, reason, err := dice.Roll(rollRequest)
	if err != nil {
		return err, ""
	}

	span.SetAttributes(rollResultValueKey.Int(result.Int()))
	span.SetAttributes(rollResultReasonKey.String(reason))

	rollResultRecorder.Record(ctx, int64(result.Int()))
	// rollQtyRecorder.Add(ctx, int64(1))
	rollQtyRecorder.Add(ctx, 1, attribute.KeyValue{Key: "initial", Value: attribute.StringValue("increment")})

	var results []string
	results = append(results, fmt.Sprintf("roll: %s", result.Description()))
	results = append(results, fmt.Sprintf("result: %d", result.Int()))
	if reason != "" {
		results = append(results, fmt.Sprintf("reason: %s", reason))
	}

	// silly memory leak here:
	requestList = append(requestList, rollRequest)

	return nil, strings.Join(results, ", ")
}
