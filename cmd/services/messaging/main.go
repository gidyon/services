package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gidyon/micro/v2"
	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/pkg/middleware/grpc/zaplogger"
	"github.com/gidyon/micro/v2/utils/encryption"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/messaging/call"
	"github.com/gidyon/services/pkg/api/subscriber"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/gidyon/services/pkg/api/messaging/sms"

	"github.com/gidyon/services/pkg/api/messaging/pusher"

	"github.com/gidyon/services/pkg/api/messaging/emailing"

	"github.com/gidyon/services/pkg/api/messaging"

	"github.com/gidyon/micro/v2/pkg/healthcheck"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	messaging_app "github.com/gidyon/services/internal/messaging"

	"github.com/gidyon/micro/v2/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/v2/pkg/middleware/grpc"
)

func main() {
	ctx := context.Background()

	// Read config
	cfg, err := config.New()
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

	// Authentication API
	authAPI, err := auth.NewAPI(&auth.Options{
		SigningKey: jwtKey,
		Issuer:     "Messaging API",
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
	app.AddEndpoint("/api/messaging/health/ready", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeReadiness,
		AutoMigrator: func() error { return nil },
	}))

	// Liveness health check
	app.AddEndpoint("/api/messaging/health/live", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
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
		emailConn, err := app.ExternalServiceConn("emailing")
		errs.Panic(err)

		pusherConn, err := app.ExternalServiceConn("pusher")
		errs.Panic(err)

		smsConn, err := app.ExternalServiceConn("sms")
		errs.Panic(err)

		callConn, err := app.ExternalServiceConn("call")
		errs.Panic(err)

		subscriberConn, err := app.ExternalServiceConn("subscriber")
		errs.Panic(err)

		app.Logger().Infoln("connected to all services")

		// Pagination hasher
		paginationHasher, err := encryption.NewHasher(string(jwtKey))
		errs.Panic(err)

		// Create messaging API instance
		messagingAPI, err := messaging_app.NewMessagingServer(ctx, &messaging_app.Options{
			SQLDBWrites:      app.GormDBByName("sqlWrites"),
			SQLDBReads:       app.GormDBByName("sqlReads"),
			Logger:           app.Logger(),
			EmailSender:      os.Getenv("SENDER_EMAIL_ADDRESS"),
			EmailClient:      emailing.NewEmailingClient(emailConn),
			PushClient:       pusher.NewPushMessagingClient(pusherConn),
			SMSClient:        sms.NewSMSAPIClient(smsConn),
			CallClient:       call.NewCallAPIClient(callConn),
			SubscriberClient: subscriber.NewSubscriberAPIClient(subscriberConn),
			AuthAPI:          authAPI,
			PaginationHasher: paginationHasher,
		})
		errs.Panic(err)

		messaging.RegisterMessagingServer(app.GRPCServer(), messagingAPI)
		errs.Panic(messaging.RegisterMessagingHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		return nil
	})
}
