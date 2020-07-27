package main

import (
	"context"
	"os"

	"github.com/gidyon/micro"
	"github.com/gidyon/micro/utils/healthcheck"

	channel_app "github.com/gidyon/services/internal/channel"

	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"
	"github.com/gidyon/services/pkg/api/channel"

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
	app.AddEndpoint("/api/channels/health/ready", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeReadiness,
		AutoMigrator: func() error { return nil },
	}))

	// Liveness health check
	app.AddEndpoint("/api/channels/health/live", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeLiveNess,
		AutoMigrator: func() error { return nil },
	}))

	app.Start(ctx, func() error {
		// Create channel tracing instance
		channelAPI, err := channel_app.NewChannelAPIServer(ctx, &channel_app.Options{
			SQLDB:         app.GormDB(),
			Logger:        app.Logger(),
			JWTSigningKey: []byte(os.Getenv("JWT_SIGNING_KEY")),
		})
		handleErr(err)

		channel.RegisterChannelAPIServer(app.GRPCServer(), channelAPI)
		handleErr(channel.RegisterChannelAPIHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		return nil
	})
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
