package main

import (
	"context"
	"os"
	"time"

	"github.com/gidyon/micro/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/gidyon/micro/v2/pkg/healthcheck"

	subscriber_app "github.com/gidyon/services/internal/subscriber"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/pkg/middleware/grpc/zaplogger"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/api/channel"
	"github.com/gidyon/services/pkg/api/subscriber"

	"github.com/gidyon/micro/v2/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/v2/pkg/middleware/grpc"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
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

	jwtKey := []byte(os.Getenv("JWT_SIGNING_KEY"))

	// Authentication API
	authAPI, err := auth.NewAPI(&auth.Options{
		SigningKey: jwtKey,
		Issuer:     "Subsriber API",
		Audience:   "users",
	})
	errs.Panic(err)

	// Generate jwt token
	token, err := authAPI.GenToken(context.Background(), &auth.Payload{Group: auth.DefaultAdminGroup()}, time.Now().Add(time.Hour*24))
	if err == nil {
		app.Logger().Infof("Test jwt is %s", token)
	}

	app.AddGRPCUnaryServerInterceptors(grpc_auth.UnaryServerInterceptor(authAPI.AuthorizeFunc))
	app.AddGRPCStreamServerInterceptors(grpc_auth.StreamServerInterceptor(authAPI.AuthorizeFunc))

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

	// Servemux option for JSON Marshaling
	app.AddServeMuxOptions(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
		},
	}))

	// Start service
	app.Start(ctx, func() error {
		// Connect to account service
		accountCC, err := app.ExternalServiceConn("account")
		errs.Panic(err)

		// Connect to channel service
		channelCC, err := app.ExternalServiceConn("channel")
		errs.Panic(err)

		app.Logger().Infoln("connected to all services")

		db := app.GormDBByName("sqlWrites")
		if os.Getenv("DB_DEBUG") != "" {
			db = db.Debug()
		}

		// Create subscriber API
		subscriberAPI, err := subscriber_app.NewSubscriberAPIServer(ctx, &subscriber_app.Options{
			SQLDB:         db,
			Logger:        app.Logger(),
			ChannelClient: channel.NewChannelAPIClient(channelCC),
			AccountClient: account.NewAccountAPIClient(accountCC),
			AuthAPI:       authAPI,
		})
		errs.Panic(err)

		subscriber.RegisterSubscriberAPIServer(app.GRPCServer(), subscriberAPI)
		errs.Panic(subscriber.RegisterSubscriberAPIHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		return nil
	})
}
