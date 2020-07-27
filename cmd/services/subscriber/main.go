package main

import (
	"context"
	"os"

	"github.com/gidyon/micro"
	"github.com/gidyon/micro/utils/healthcheck"

	subscriber_app "github.com/gidyon/services/internal/subscriber"

	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/api/channel"
	"github.com/gidyon/services/pkg/api/subscriber"

	"github.com/gidyon/micro/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"
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
	app.AddEndpoint("/api/subscribers/health/ready", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeReadiness,
		AutoMigrator: func() error { return nil },
	}))

	// Liveness health check
	app.AddEndpoint("/api/subscribers/health/live", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeLiveNess,
		AutoMigrator: func() error { return nil },
	}))

	// Start service
	app.Start(ctx, func() error {
		// Connect to account service
		accountCC, err := app.DialExternalService(ctx, "account")
		handleErr(err)
		app.Logger().Infoln("connected to account service")

		// Connect to channel service
		channelCC, err := app.DialExternalService(ctx, "channel")
		handleErr(err)
		app.Logger().Infoln("connected to channel service")

		app.Logger().Infoln("connected to all services")

		// Create subscriber API
		subscriberAPI, err := subscriber_app.NewSubscriberAPIServer(ctx, &subscriber_app.Options{
			SQLDB:         app.GormDB(),
			Logger:        app.Logger(),
			ChannelClient: channel.NewChannelAPIClient(channelCC),
			AccountClient: account.NewAccountAPIClient(accountCC),
			JWTSigningKey: []byte(os.Getenv("JWT_SIGNING_KEY")),
		})
		handleErr(err)

		subscriber.RegisterSubscriberAPIServer(app.GRPCServer(), subscriberAPI)
		handleErr(subscriber.RegisterSubscriberAPIHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		return nil
	})
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
