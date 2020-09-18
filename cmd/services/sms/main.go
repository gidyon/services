package main

import (
	"context"
	"os"

	"github.com/gidyon/micro"
	"github.com/gidyon/micro/utils/healthcheck"

	sms_app "github.com/gidyon/services/internal/messaging/sms"

	"github.com/gidyon/services/pkg/api/messaging/sms"
	"github.com/gidyon/services/pkg/utils/errs"

	"github.com/gidyon/micro/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"
)

func main() {
	ctx := context.Background()

	// Read config
	cfg, err := config.New(config.FromFile)
	errs.Panic(err)

	// Create service instance
	app, err := micro.NewService(ctx, cfg, nil)
	errs.Panic(err)

	// Recovery middleware
	recoveryUIs, recoverySIs := app_grpc_middleware.AddRecovery()
	app.AddGRPCUnaryServerInterceptors(recoveryUIs...)
	app.AddGRPCStreamServerInterceptors(recoverySIs...)

	// Readiness health check
	app.AddEndpoint("/api/sms/health/ready", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeReadiness,
		AutoMigrator: func() error { return nil },
	}))

	// Liveness health check
	app.AddEndpoint("/api/sms/health/live", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeLiveNess,
		AutoMigrator: func() error { return nil },
	}))

	// Start service
	app.Start(ctx, func() error {
		// Create sms API instance
		smsAPI, err := sms_app.NewSMSAPIServer(ctx, &sms_app.Options{
			Logger:        app.Logger(),
			JWTSigningKey: []byte(os.Getenv("JWT_SIGNING_KEY")),
			APIKey:        os.Getenv("SMS_API_KEY"),
			AuthToken:     os.Getenv("SMS_AUTH_TOKEN"),
			APIUsername:   os.Getenv("SMS_API_USERNAME"),
			APIPassword:   os.Getenv("SMS_API_PASSWORD"),
			APIURL:        os.Getenv("SMS_API_URL"),
		})
		errs.Panic(err)

		sms.RegisterSMSAPIServer(app.GRPCServer(), smsAPI)
		sms.RegisterSMSAPIHandler(ctx, app.RuntimeMux(), app.ClientConn())

		return nil
	})
}
