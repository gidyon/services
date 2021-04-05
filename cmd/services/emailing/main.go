package main

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	emailing_service "github.com/gidyon/services/internal/messaging/emailing"
	"github.com/gidyon/services/pkg/api/messaging/emailing"

	"github.com/gidyon/micro/v2"
	"github.com/gidyon/micro/v2/pkg/healthcheck"
	"github.com/gidyon/micro/v2/pkg/middleware/grpc/zaplogger"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"

	"github.com/gidyon/micro/v2/pkg/config"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func main() {
	ctx := context.Background()

	// Read config
	cfg, err := config.New()
	errs.Panic(err)

	// initialize logger
	errs.Panic(zaplogger.Init(cfg.LogLevel(), ""))

	zaplogger.Log = zaplogger.Log.WithOptions(zap.WithCaller(true))

	serviceLogger := zaplogger.ZapGrpcLoggerV2(zaplogger.Log)

	service, err := micro.NewService(ctx, cfg, serviceLogger)
	errs.Panic(err)

	// Recovery middleware
	recoveryUIs, recoverySIs := app_grpc_middleware.AddRecovery()
	service.AddGRPCUnaryServerInterceptors(recoveryUIs...)
	service.AddGRPCStreamServerInterceptors(recoverySIs...)

	// Logging middleware
	logginUIs, loggingSIs := app_grpc_middleware.AddLogging(zaplogger.Log)
	service.AddGRPCUnaryServerInterceptors(logginUIs...)
	service.AddGRPCStreamServerInterceptors(loggingSIs...)

	jwtKey := []byte(strings.TrimSpace(os.Getenv("JWT_SIGNING_KEY")))

	if len(jwtKey) == 0 {
		errs.Panic(errors.New("missing jwt key"))
	}

	// Authentication API
	authAPI, err := auth.NewAPI(&auth.Options{
		SigningKey: jwtKey,
		Issuer:     "Emailing API",
		Audience:   "users",
	})
	errs.Panic(err)

	// Generate jwt token
	token, err := authAPI.GenToken(context.Background(), &auth.Payload{Group: auth.DefaultAdminGroup()}, time.Now().Add(time.Hour*24))
	if err == nil {
		service.Logger().Infof("Test jwt is %s", token)
	}

	service.AddGRPCUnaryServerInterceptors(grpc_auth.UnaryServerInterceptor(authAPI.AuthorizeFunc))
	service.AddGRPCStreamServerInterceptors(grpc_auth.StreamServerInterceptor(authAPI.AuthorizeFunc))

	// Readiness health check
	service.AddEndpoint("/api/emailing/health/ready", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      service,
		Type:         healthcheck.ProbeReadiness,
		AutoMigrator: func() error { return nil },
	}))

	// Liveness health check
	service.AddEndpoint("/api/emailing/health/live", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      service,
		Type:         healthcheck.ProbeLiveNess,
		AutoMigrator: func() error { return nil },
	}))

	// Servemux option for JSON Marshaling
	service.AddServeMuxOptions(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
		},
	}))

	// Start service
	service.Start(ctx, func() error {
		port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
		errs.Panic(err)

		// Create emailing API
		emailingAPI, err := emailing_service.NewEmailingAPIServer(ctx, &emailing_service.Options{
			AuthAPI:      authAPI,
			Logger:       service.Logger(),
			SMTPHost:     os.Getenv("SMTP_HOST"),
			SMTPPort:     port,
			SMTPUsername: os.Getenv("SMTP_USERNAME"),
			SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		})
		errs.Panic(err)

		emailing.RegisterEmailingServer(service.GRPCServer(), emailingAPI)
		errs.Panic(emailing.RegisterEmailingHandler(ctx, service.RuntimeMux(), service.ClientConn()))

		return nil
	})
}
