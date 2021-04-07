package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gidyon/micro/v2/pkg/conn"
	"github.com/gidyon/micro/v2/utils/mdutil"
	"github.com/gidyon/services/pkg/api/account"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func waitForReadyInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	return invoker(ctx, method, req, reply, cc, append(opts, grpc.WaitForReady(true))...)
}

func main() {
	// Generate localhost certificate
	creds, err := credentials.NewClientTLSFromFile("/home/gideon/go/src/github.com/gidyon/services/certs/localhost/cert.pem", "")
	handleErr(err)

	dopts := []grpc.DialOption{
		// grpc.WithInsecure(),
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [ { "round_robin": {} } ] }`),
		// Load balancer scheme
		grpc.WithDisableServiceConfig(),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				waitForReadyInterceptor,
			),
		)}

	ctx := context.Background()

	fmt.Println("dialing")

	conn, err := conn.DialService(ctx, &conn.GRPCDialOptions{
		ServiceName: "self",
		Address:     "localhost:9003",
		DialOptions: dopts,
	})
	handleErr(err)

	// conn, err := grpc.Dial("localhost:8080", dopts...)
	// handleErr(err)

	cc := account.NewAccountAPIClient(conn)

	md := metadata.Pairs("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6IiIsIlByb2plY3RJRCI6IiIsIk5hbWVzIjoiIiwiUGhvbmVOdW1iZXIiOiIiLCJFbWFpbEFkZHJlc3MiOiIiLCJHcm91cCI6IiIsImF1ZCI6ImFjY291bnRzIiwiZXhwIjoxNjI2Mzc0MTUzLCJpc3MiOiJBY2NvdW50cyBBUEkifQ.k7QQAvHm_Sn4E8Fm1VbrkmeMyE-kkUNMhW8f8RVoDNM")

	ctx2 := mdutil.AddMD(ctx, md)

	header := grpc.Header(&md)

	fmt.Println("getting account")

	res, err := cc.GetAccount(ctx2, &account.GetAccountRequest{
		AccountId: "1",
	}, grpc.WaitForReady(true), header)
	handleErr(err)

	log.Println("done getting account")

	log.Println(res)
}
