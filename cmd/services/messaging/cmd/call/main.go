package main

import (
	"context"
	"os"

	"github.com/gidyon/micro"
	"github.com/gidyon/micro/utils/healthcheck"

	call_app "github.com/gidyon/services/internal/messaging/call"

	"github.com/gidyon/services/pkg/api/messaging/call"

	"github.com/gidyon/micro/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"
)

func main() {
	ctx := context.Background()

	// Read config
	cfg, err := config.New(config.FromFile)
	handleErr(err)

	// Create service
	callSrv, err := micro.NewService(ctx, cfg, nil)
	handleErr(err)

	// Recovery middleware
	recoveryUIs, recoverySIs := app_grpc_middleware.AddRecovery()
	callSrv.AddGRPCUnaryServerInterceptors(recoveryUIs...)
	callSrv.AddGRPCStreamServerInterceptors(recoverySIs...)

	// Readiness health check
	callSrv.AddEndpoint("/api/messaging/call/health/ready", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      callSrv,
		Type:         healthcheck.ProbeReadiness,
		AutoMigrator: func() error { return nil },
	}))

	// Liveness health check
	callSrv.AddEndpoint("/api/messaging/call/health/live", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      callSrv,
		Type:         healthcheck.ProbeLiveNess,
		AutoMigrator: func() error { return nil },
	}))

	// Start the service
	callSrv.Start(ctx, func() error {
		callAPI, err := call_app.NewCallAPIServer(ctx, &call_app.Options{
			Logger:        callSrv.Logger(),
			JWTSigningKey: []byte(os.Getenv("JWT_SIGNING_KEY")),
		})
		handleErr(err)

		call.RegisterCallAPIServer(callSrv.GRPCServer(), callAPI)

		return nil
	})
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
