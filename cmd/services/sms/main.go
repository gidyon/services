package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gidyon/micro/v2"
	"github.com/gidyon/micro/v2/pkg/healthcheck"
	"github.com/gidyon/micro/v2/pkg/middleware/grpc/zaplogger"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	sms_app "github.com/gidyon/services/internal/messaging/sms"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/messaging/sms"

	"github.com/gidyon/micro/v2/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/v2/pkg/middleware/grpc"
)

func main() {
	ctx := context.Background()

	// Read config
	cfg, err := config.New(config.FromFile)
	errs.Panic(err)

	// initialize logger
	errs.Panic(zaplogger.Init(cfg.LogLevel(), ""))

	zaplogger.Log = zaplogger.Log.WithOptions(zap.WithCaller(true))

	appLogger := zaplogger.ZapGrpcLoggerV2(zaplogger.Log)

	app, err := micro.NewService(ctx, cfg, appLogger)
	errs.Panic(err)

	// Recovery middleware
	recoveryUIs, recoverySIs := app_grpc_middleware.AddRecovery()
	app.AddGRPCUnaryServerInterceptors(recoveryUIs...)
	app.AddGRPCStreamServerInterceptors(recoverySIs...)

	// Logging middleware
	logginUIs, loggingSIs := app_grpc_middleware.AddLogging(zaplogger.Log)
	app.AddGRPCUnaryServerInterceptors(logginUIs...)
	app.AddGRPCStreamServerInterceptors(loggingSIs...)

	jwtKey := []byte(strings.TrimSpace(os.Getenv("JWT_SIGNING_KEY")))

	if len(jwtKey) == 0 {
		errs.Panic(errors.New("missing jwt key"))
	}

	authAPI, err := auth.NewAPI(&auth.Options{
		SigningKey: jwtKey,
		Issuer:     "SMS API",
		Audience:   "accounts",
	})
	errs.Panic(err)

	// Generate jwt token
	token, err := authAPI.GenToken(context.Background(), &auth.Payload{}, time.Now().Add(time.Hour*24))
	if err == nil {
		app.Logger().Infof("test jwt is [%s]", token)
	}

	app.AddGRPCUnaryServerInterceptors(grpc_auth.UnaryServerInterceptor(authAPI.AuthorizeFunc))
	app.AddGRPCStreamServerInterceptors(grpc_auth.StreamServerInterceptor(authAPI.AuthorizeFunc))

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

	// Servemux option for JSON Marshaling
	app.AddServeMuxOptions(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
		},
	}))

	// Start service
	app.Start(ctx, func() error {
		// Create sms API instance
		smsAPI, err := sms_app.NewSMSAPIServer(ctx, &sms_app.Options{
			Logger:  app.Logger(),
			AuthAPI: authAPI,
			SmsAuth: sms.SendSMSRequest_SMSAuth{
				ApiUrl:    os.Getenv("SMS_API_URL"),
				SenderId:  os.Getenv("SENDER_ID"),
				ApiKey:    os.Getenv("SMS_API_KEY"),
				ClientId:  os.Getenv("SMS_API_USERNAME"),
				AuthToken: os.Getenv("SMS_AUTH_TOKEN"),
				AccessKey: os.Getenv("SMS_ACCESS_TOKEN"),
			},
		})
		errs.Panic(err)

		sms.RegisterSMSAPIServer(app.GRPCServer(), smsAPI)
		errs.Panic(sms.RegisterSMSAPIHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		return nil
	})
}
