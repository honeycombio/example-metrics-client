package main

import (
	"fmt"
	"net/http"

	"github.com/justinian/dice"
)

func hello(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/html")

	if req.URL.Path == "/" {
		fmt.Fprintf(w, "<h1>Welcome to polyhedron!</h1><p>Try <a href=\"/1d6\">/1d6</a></p>")
		return
	}

	result, reason, err := dice.Roll(req.URL.Path)
	if err != nil {
		fmt.Fprintf(w, "Error: %v\n", err)
		return
	}

	fmt.Fprintf(w, "<b>Roll:</b> %s<br /><b>Result:</b> %d<br />%s", result.Description(), result.Int(), reason)
}

func main() {
	http.HandleFunc("/", hello)
	http.ListenAndServe(":8090", nil)
}
