package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/justinian/dice"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("polyhedron")
var errorKey = attribute.Key("error")

func serve() {
	http.Handle("/", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
	}), "handler"))
	if err := http.ListenAndServe(":8090", nil); err != nil {
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

	var results []string
	results = append(results, fmt.Sprintf("roll: %s", result.Description()))
	results = append(results, fmt.Sprintf("result: %d", result.Int()))
	if reason != "" {
		results = append(results, fmt.Sprintf("reason: %s", reason))
	}

	return nil, strings.Join(results, ", ")
}
