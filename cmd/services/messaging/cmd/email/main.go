package main

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/gidyon/micro"
	"github.com/gidyon/micro/utils/healthcheck"

	emailing_app "github.com/gidyon/services/internal/messaging/emailing"

	"github.com/gidyon/services/pkg/api/messaging/emailing"

	"github.com/gidyon/micro/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"
)

func main() {
	ctx := context.Background()

	// Read config
	cfg, err := config.New()
	handleErr(err)

	// Create service instance
	app, err := micro.NewService(ctx, cfg, nil)
	handleErr(err)

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
		handleErr(err)

		// Create emailing API
		emailingAPI, err := emailing_app.NewEmailingAPIServer(ctx, &emailing_app.Options{
			Logger:        app.Logger(),
			JWTSigningKey: []byte(os.Getenv("JWT_SIGNING_KEY")),
			SMTPHost:      os.Getenv("SMTP_HOST"),
			SMTPPort:      port,
			SMTPUsername:  os.Getenv("SMTP_USERNAME"),
			SMTPPassword:  os.Getenv("SMTP_PASSWORD"),
		})
		handleErr(err)

		emailing.RegisterEmailingServer(app.GRPCServer(), emailingAPI)

		return nil
	})
}

func setIfempty(val1, val2 string, swap ...bool) string {
	if len(swap) > 0 && swap[0] {
		if strings.TrimSpace(val2) == "" {
			return val1
		}
		return val2
	}
	if strings.TrimSpace(val1) == "" {
		return val2
	}
	return val1
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
