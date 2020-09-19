package main

import (
	"flag"
	"log"
	"net/http"
)

var (
	port     = flag.String("port", ":9090", "Port to listen to")
	certFile = flag.String("cert", "/Users/jessegitaka/go/src/github.com/gidyon/services/certs/localhost/cert.pem", "PKI Public Key file")
	keyFile  = flag.String("key", "/Users/jessegitaka/go/src/github.com/gidyon/services/certs/localhost/key.pem", "PKI Private Key file")
	insecure = flag.Bool("insecure", false, "Whether to use insecure http")
)

func main() {
	flag.Parse()

	handler := http.FileServer(http.Dir("./dist/"))

	log.Printf("API Documentation server started on port %s\n", *port)

	if *insecure {
		log.Fatalln(http.ListenAndServe(*port, handler))
		return
	}
	log.Fatalln(http.ListenAndServeTLS(*port, *certFile, *keyFile, handler))
}
