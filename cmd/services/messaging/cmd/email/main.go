package main

import (
	"context"
	"os"
	"strconv"

	"github.com/gidyon/micro"
	"github.com/gidyon/micro/utils/healthcheck"

	emailing_app "github.com/gidyon/services/internal/messaging/emailing"

	"github.com/gidyon/services/pkg/api/messaging/emailing"
	"github.com/gidyon/services/pkg/utils/errs"

	"github.com/gidyon/micro/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"
)

func main() {
	ctx := context.Background()

	// Read config
	cfg, err := config.New()
	errs.Panic(err)

	// Create service instance
	app, err := micro.NewService(ctx, cfg, nil)
	errs.Panic(err)

	// Recovery middleware
	recoveryUIs, recoverySIs := app_grpc_middleware.AddRecovery()
	app.AddGRPCUnaryServerInterceptors(recoveryUIs...)
	app.AddGRPCStreamServerInterceptors(recoverySIs...)

	// Readiness health check
	app.AddEndpoint("/api/messaging/emailing/health/ready", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeReadiness,
		AutoMigrator: func() error { return nil },
	}))

	// Liveness health check
	app.AddEndpoint("/api/messaging/emailing/health/live", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeLiveNess,
		AutoMigrator: func() error { return nil },
	}))

	// Start service
	app.Start(ctx, func() error {
		port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
		errs.Panic(err)

		// Create emailing API
		emailingAPI, err := emailing_app.NewEmailingAPIServer(ctx, &emailing_app.Options{
			Logger:        app.Logger(),
			JWTSigningKey: []byte(os.Getenv("JWT_SIGNING_KEY")),
			SMTPHost:      os.Getenv("SMTP_HOST"),
			SMTPPort:      port,
			SMTPUsername:  os.Getenv("SMTP_USERNAME"),
			SMTPPassword:  os.Getenv("SMTP_PASSWORD"),
		})
		errs.Panic(err)

		emailing.RegisterEmailingServer(app.GRPCServer(), emailingAPI)

		return nil
	})
}
