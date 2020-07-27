package main

import (
	"context"
	"os"

	"github.com/gidyon/micro"
	"github.com/gidyon/services/pkg/api/messaging/call"
	"github.com/gidyon/services/pkg/api/subscriber"

	"github.com/gidyon/services/pkg/api/messaging/sms"

	"github.com/gidyon/services/pkg/api/messaging/push"

	"github.com/gidyon/services/pkg/api/messaging/emailing"

	"github.com/gidyon/services/pkg/api/messaging"

	"google.golang.org/grpc"

	"github.com/gidyon/micro/utils/healthcheck"

	messaging_app "github.com/gidyon/services/internal/messaging"

	"github.com/gidyon/micro/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"
)

func main() {
	ctx := context.Background()

	// Read config
	cfg, err := config.New()
	handleErr(err)

	// Create service toolkit
	app, err := micro.NewService(ctx, cfg, nil)
	handleErr(err)

	// Recovery middleware
	recoveryUIs, recoverySIs := app_grpc_middleware.AddRecovery()
	app.AddGRPCUnaryServerInterceptors(recoveryUIs...)
	app.AddGRPCStreamServerInterceptors(recoverySIs...)

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

	// Start service
	app.Start(ctx, func() error {
		emailConn, err := app.DialExternalService(ctx, "emailing", grpc.WithBlock())
		handleErr(err)
		app.Logger().Infoln("connected to emailing service")

		pusherConn, err := app.DialExternalService(ctx, "pusher", grpc.WithBlock())
		handleErr(err)
		app.Logger().Infoln("connected to pusher service")

		smsConn, err := app.DialExternalService(ctx, "sms", grpc.WithBlock())
		handleErr(err)
		app.Logger().Infoln("connected to sms service")

		callConn, err := app.DialExternalService(ctx, "call", grpc.WithBlock())
		handleErr(err)
		app.Logger().Infoln("connected to call service")

		subscriberConn, err := app.DialExternalService(ctx, "subscriber", grpc.WithBlock())
		handleErr(err)
		app.Logger().Infoln("connected to subscriber service")

		app.Logger().Infoln("connected to all services")

		// Create messaging API instance
		messagingAPI, err := messaging_app.NewMessagingServer(ctx, &messaging_app.Options{
			SQLDB:            app.GormDB(),
			Logger:           app.Logger(),
			JWTSigningKey:    []byte(os.Getenv("JWT_SIGNING_KEY")),
			EmailSender:      os.Getenv("SENDER_EMAIL_ADDRESS"),
			EmailClient:      emailing.NewEmailingClient(emailConn),
			PushClient:       push.NewPushMessagingClient(pusherConn),
			SMSClient:        sms.NewSMSAPIClient(smsConn),
			CallClient:       call.NewCallAPIClient(callConn),
			SubscriberClient: subscriber.NewSubscriberAPIClient(subscriberConn),
		})
		handleErr(err)

		messaging.RegisterMessagingServer(app.GRPCServer(), messagingAPI)
		handleErr(messaging.RegisterMessagingHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		return nil
	})
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
