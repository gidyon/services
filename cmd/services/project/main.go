package main

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/gorm"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/gidyon/micro/v2"
	"github.com/gidyon/micro/v2/pkg/middleware/grpc/zaplogger"

	"github.com/gidyon/micro/v2/pkg/healthcheck"

	project_app_v1 "github.com/gidyon/services/internal/project/v1"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
	project_api_v1 "github.com/gidyon/services/pkg/api/project/v1"

	"github.com/gidyon/micro/v2/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/v2/pkg/middleware/grpc"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
)

func main() {
	ctx := context.Background()

	fmt.Println(resolver.GetDefaultScheme())

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

	// Payload interceptor
	alwaysLoggingDeciderServer := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return true }
	app.AddGRPCUnaryServerInterceptors(grpc_zap.PayloadUnaryServerInterceptor(zaplogger.Log, alwaysLoggingDeciderServer))
	app.AddGRPCStreamServerInterceptors(grpc_zap.PayloadStreamServerInterceptor(zaplogger.Log, alwaysLoggingDeciderServer))

	jwtKey := []byte(os.Getenv("JWT_SIGNING_KEY"))

	// Authentication API
	authAPI, err := auth.NewAPI(&auth.Options{
		SigningKey: jwtKey,
		Issuer:     "Projects API",
		Audience:   "consumers",
	})
	errs.Panic(err)

	app.AddGRPCUnaryServerInterceptors(grpc_auth.UnaryServerInterceptor(authAPI.AuthorizeFunc))
	app.AddGRPCStreamServerInterceptors(grpc_auth.StreamServerInterceptor(authAPI.AuthorizeFunc))

	// Readiness health check
	app.AddEndpoint("/v1/projects:appReady", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeReadiness,
		AutoMigrator: func() error { return nil },
	}))

	// Liveness health check
	app.AddEndpoint("/v1/projects:appLive", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
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

	// Bootstrapping service
	app.Start(ctx, func() error {
		projectAPI, err := project_app_v1.NewProjectAPI(ctx, &project_app_v1.Options{
			AuthAPI: authAPI,
			SqlDb: func() *gorm.DB {
				if os.Getenv("DB_DEBUG") != "" {
					return app.GormDBByName("sqlWrites").Debug()
				}
				return app.GormDBByName("sqlWrites")
			}(),
			Logger: app.Logger(),
		})
		errs.Panic(err)

		project_api_v1.RegisterProjectAPIServer(app.GRPCServer(), projectAPI)
		errs.Panic(project_api_v1.RegisterProjectAPIHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		return nil
	})
}
