package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/gidyon/micro"
	"github.com/gidyon/micro/pkg/healthcheck"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"

	emailing_app "github.com/gidyon/services/internal/messaging/emailing"

	"github.com/gidyon/micro/pkg/grpc/auth"
	"github.com/gidyon/micro/utils/errs"
	"github.com/gidyon/services/pkg/api/messaging/emailing"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

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

	jwtKey := []byte(os.Getenv("JWT_SIGNING_KEY"))

	// Authentication API
	authAPI, err := auth.NewAPI(jwtKey, "USSD Log API", "users")
	errs.Panic(err)

	// Generate jwt token
	token, err := authAPI.GenToken(context.Background(), &auth.Payload{Group: auth.AdminGroup()}, time.Now().Add(time.Hour*24))
	if err == nil {
		app.Logger().Infof("Test jwt is %s", token)
	}

	app.AddGRPCUnaryServerInterceptors(grpc_auth.UnaryServerInterceptor(authAPI.AuthFunc))
	app.AddGRPCStreamServerInterceptors(grpc_auth.StreamServerInterceptor(authAPI.AuthFunc))

	// Readiness health check
	app.AddEndpoint("/api/emailing/health/ready", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeReadiness,
		AutoMigrator: func() error { return nil },
	}))

	// Liveness health check
	app.AddEndpoint("/api/emailing/health/live", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeLiveNess,
		AutoMigrator: func() error { return nil },
	}))

	// Servemux option for JSON Marshaling
	app.AddServeMuxOptions(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
		},
	}))

	// Start service
	app.Start(ctx, func() error {
		port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
		errs.Panic(err)

		// Create emailing API
		emailingAPI, err := emailing_app.NewEmailingAPIServer(ctx, &emailing_app.Options{
			AuthAPI:      authAPI,
			Logger:       app.Logger(),
			SMTPHost:     os.Getenv("SMTP_HOST"),
			SMTPPort:     port,
			SMTPUsername: os.Getenv("SMTP_USERNAME"),
			SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		})
		errs.Panic(err)

		emailing.RegisterEmailingServer(app.GRPCServer(), emailingAPI)
		emailing.RegisterEmailingHandler(ctx, app.RuntimeMux(), app.ClientConn())

		return nil
	})
}
