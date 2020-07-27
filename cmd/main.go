package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	})
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	})
	http.ListenAndServe(":7070", mux)
}

// {"level":"info","ts":"2020-03-08T11:35:27+03:00","msg":"pickfirstBalancer: HandleSubConnStateChange: 0xc0002ac9d0, {TRANSIENT_FAILURE connection error: desc = \"transport: Error while dialing dial tcp 127.0.0.1:9050: connect: connection refused\"}","system":"grpc","grpc_log":true}
