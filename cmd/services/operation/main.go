package main

import (
	"context"
	"os"

	"github.com/gidyon/micro"
	"github.com/gidyon/micro/utils/healthcheck"

	operation_app "github.com/gidyon/services/internal/operation"

	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"
	"github.com/gidyon/services/pkg/api/operation"

	"github.com/gidyon/micro/pkg/config"
)

func main() {
	ctx := context.Background()

	// Read config
	cfg, err := config.New(config.FromFile)
	handleErr(err)

	// Create service
	app, err := micro.NewService(ctx, cfg, nil)
	handleErr(err)

	// Add middlewares
	recoveryUIs, recoverySIs := app_grpc_middleware.AddRecovery()
	app.AddGRPCUnaryServerInterceptors(recoveryUIs...)
	app.AddGRPCStreamServerInterceptors(recoverySIs...)

	// Readiness health check
	app.AddEndpoint("/health/ready", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeReadiness,
		AutoMigrator: func() error { return nil },
	}))

	// Liveness health check
	app.AddEndpoint("/health/live", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeLiveNess,
		AutoMigrator: func() error { return nil },
	}))

	// Start the service
	app.Start(ctx, func() error {
		operationAPI, err := operation_app.NewOperationAPIService(ctx, &operation_app.Options{
			RedisClient:   app.RedisClient(),
			Logger:        app.Logger(),
			JWTSigningKey: []byte(os.Getenv("JWT_SIGNING_KEY")),
		})
		handleErr(err)

		operation.RegisterOperationAPIServer(app.GRPCServer(), operationAPI)

		return nil
	})
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
