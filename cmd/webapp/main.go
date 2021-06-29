package main

import (
	"flag"
	"log"
	"net/http"
)

var (
	port = flag.String("port", ":80", "Port to listen format is :port")
)

func main() {
	flag.Parse()

	// API documentation
	http.Handle("/", http.StripPrefix("/documentation/", http.FileServer(http.Dir("./apidoc"))))

	log.Printf("Static file server started on port %s\n", *port)

	log.Fatalln(http.ListenAndServe(*port, nil))
}
