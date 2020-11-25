package main

import (
	"context"
	"log"

	"google.golang.org/grpc/metadata"

	"github.com/gidyon/services/pkg/api/account"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// Generate localhost certificate
	// _, clientTLS, err := tlslocalhost.GenCerts()
	// handleErr(err)

	// creds := credentials.NewClientTLSFromCert(clientTLS.RootCAs, "")
	creds, err := credentials.NewClientTLSFromFile("/home/gideon/go/src/github.com/gidyon/services/certs/localhost/cert.pem", "")
	handleErr(err)

	dopts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}

	ctx := context.Background()

	conn, err := grpc.Dial("localhost:7070", dopts...)
	handleErr(err)

	cc := account.NewAccountAPIClient(conn)

	md := metadata.MD{}

	header := grpc.Header(&md)

	res, err := cc.GetAccount(ctx, &account.GetAccountRequest{
		AccountId: "1",
	}, grpc.WaitForReady(true), header)
	handleErr(err)

	log.Println(md)
	log.Println("done getting account")

	log.Println(res)
}
