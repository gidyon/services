package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gidyon/micro/pkg/conn"
	"github.com/gidyon/micro/utils/mdutil"
	"github.com/gidyon/services/pkg/api/channel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("calling remote conn")
	tlsCreds, err := credentials.NewClientTLSFromFile("/home/gideon/go/src/github.com/gidyon/services/certs/localhost/cert.pem", "localhost")
	handleErr(err)

	cc, err := conn.DialService(ctx, &conn.GRPCDialOptions{
		K8Service:   false,
		Address:     "localhost:9080",
		DialOptions: []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds), grpc.WithBlock()},
	})
	// cc, err := grpc.Dial("dns:///localhost:9080")
	handleErr(err)

	client := channel.NewChannelAPIClient(cc)

	ctx = mdutil.AddMD(ctx, metadata.Pairs("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6IjEiLCJOYW1lcyI6IkJvbmVkdXN0IiwiUGhvbmVOdW1iZXIiOiIrOTk0IDUgNTAgMjcyMzUgOTYgNyIsIkVtYWlsQWRkcmVzcyI6IiIsIkdyb3VwIjoiQURNSU4iLCJpc3MiOiJlbXJzIn0.Nlg5z2cfFAmr5eSTixiVizlahKcfdn_0UaJWKbdLLWU"))

	log.Println("calling list channels")

	listRes, err := client.ListChannels(ctx, &channel.ListChannelsRequest{
		PageSize: 3,
	}, grpc.WaitForReady(true))
	handleErr(err)

	for _, channelPB := range listRes.Channels {
		fmt.Printf("channel name: %v\n", channelPB.Title)
	}
}
