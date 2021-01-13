package main

import (
	"context"
	"os"
	"time"

	"github.com/gidyon/micro"
	"github.com/gidyon/micro/pkg/healthcheck"
	"github.com/gidyon/micro/v2/pkg/middleware/grpc/zaplogger"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	call_app "github.com/gidyon/services/internal/messaging/call"

	"github.com/gidyon/micro/pkg/grpc/auth"
	"github.com/gidyon/micro/utils/errs"
	"github.com/gidyon/services/pkg/api/messaging/call"

	"github.com/gidyon/micro/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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

	callSrv, err := micro.NewService(ctx, cfg, appLogger)
	errs.Panic(err)

	// Recovery middleware
	recoveryUIs, recoverySIs := app_grpc_middleware.AddRecovery()
	callSrv.AddGRPCUnaryServerInterceptors(recoveryUIs...)
	callSrv.AddGRPCStreamServerInterceptors(recoverySIs...)

	// Logging middleware
	logginUIs, loggingSIs := app_grpc_middleware.AddLogging(zaplogger.Log)
	callSrv.AddGRPCUnaryServerInterceptors(logginUIs...)
	callSrv.AddGRPCStreamServerInterceptors(loggingSIs...)

	jwtKey := []byte(os.Getenv("JWT_SIGNING_KEY"))

	// Authentication API
	authAPI, err := auth.NewAPI(jwtKey, "USSD Log API", "users")
	errs.Panic(err)

	// Generate jwt token
	token, err := authAPI.GenToken(context.Background(), &auth.Payload{Group: auth.AdminGroup()}, time.Now().Add(time.Hour*24))
	if err == nil {
		callSrv.Logger().Infof("Test jwt is %s", token)
	}

	callSrv.AddGRPCUnaryServerInterceptors(grpc_auth.UnaryServerInterceptor(authAPI.AuthFunc))
	callSrv.AddGRPCStreamServerInterceptors(grpc_auth.StreamServerInterceptor(authAPI.AuthFunc))

	// Readiness health check
	callSrv.AddEndpoint("/api/calls/health/ready", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      callSrv,
		Type:         healthcheck.ProbeReadiness,
		AutoMigrator: func() error { return nil },
	}))

	// Liveness health check
	callSrv.AddEndpoint("/api/calls/health/live", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      callSrv,
		Type:         healthcheck.ProbeLiveNess,
		AutoMigrator: func() error { return nil },
	}))

	// Servemux option for JSON Marshaling
	callSrv.AddServeMuxOptions(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
		},
	}))

	// Start the service
	callSrv.Start(ctx, func() error {
		callAPI, err := call_app.NewCallAPIServer(ctx, &call_app.Options{
			Logger:  callSrv.Logger(),
			AuthAPI: authAPI,
		})
		errs.Panic(err)

		call.RegisterCallAPIServer(callSrv.GRPCServer(), callAPI)
		call.RegisterCallAPIHandler(ctx, callSrv.RuntimeMux(), callSrv.ClientConn())

		return nil
	})
}
