package main

import (
	"context"
	"os"

	"github.com/gidyon/micro"
	httpmiddleware "github.com/gidyon/micro/pkg/http"
	"github.com/gidyon/services/pkg/api/messaging/call"
	"github.com/gidyon/services/pkg/api/subscriber"
	"github.com/gidyon/services/pkg/auth"
	"github.com/gidyon/services/pkg/utils/encryption"
	"github.com/gidyon/services/pkg/utils/errs"
	"github.com/gorilla/securecookie"

	"github.com/gidyon/services/pkg/api/messaging/sms"

	"github.com/gidyon/services/pkg/api/messaging/pusher"

	"github.com/gidyon/services/pkg/api/messaging/emailing"

	"github.com/gidyon/services/pkg/api/messaging"

	"github.com/gidyon/micro/utils/healthcheck"

	messaging_app "github.com/gidyon/services/internal/messaging"

	"github.com/gidyon/micro/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"
)

func main() {
	ctx := context.Background()

	apiHashKey, err := encryption.ParseKey([]byte(os.Getenv("API_HASH_KEY")))
	errs.Panic(err)

	apiBlockKey, err := encryption.ParseKey([]byte(os.Getenv("API_BLOCK_KEY")))
	errs.Panic(err)

	// Read config
	cfg, err := config.New()
	errs.Panic(err)

	// Create service toolkit
	app, err := micro.NewService(ctx, cfg, nil)
	errs.Panic(err)

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

	sc := securecookie.New(apiHashKey, apiBlockKey)

	// Cookie based authentication
	app.AddHTTPMiddlewares(httpmiddleware.CookieToJWTMiddleware(&httpmiddleware.CookieJWTOptions{
		SecureCookie: sc,
		AuthHeader:   auth.Header(),
		AuthScheme:   auth.Scheme(),
		CookieName:   auth.JWTCookie(),
	}))

	// Start service
	app.Start(ctx, func() error {
		emailConn, err := app.DialExternalService(ctx, "emailing")
		errs.Panic(err)
		app.Logger().Infoln("connected to emailing service")

		pusherConn, err := app.DialExternalService(ctx, "pusher")
		errs.Panic(err)
		app.Logger().Infoln("connected to pusher service")

		smsConn, err := app.DialExternalService(ctx, "sms")
		errs.Panic(err)
		app.Logger().Infoln("connected to sms service")

		callConn, err := app.DialExternalService(ctx, "call")
		errs.Panic(err)
		app.Logger().Infoln("connected to call service")

		subscriberConn, err := app.DialExternalService(ctx, "subscriber")
		errs.Panic(err)
		app.Logger().Infoln("connected to subscriber service")

		app.Logger().Infoln("connected to all services")

		// Create messaging API instance
		messagingAPI, err := messaging_app.NewMessagingServer(ctx, &messaging_app.Options{
			SQLDBWrites:      app.GormDBByName("sqlWrites"),
			SQLDBReads:       app.GormDBByName("sqlReads"),
			Logger:           app.Logger(),
			JWTSigningKey:    []byte(os.Getenv("JWT_SIGNING_KEY")),
			EmailSender:      os.Getenv("SENDER_EMAIL_ADDRESS"),
			EmailClient:      emailing.NewEmailingClient(emailConn),
			PushClient:       pusher.NewPushMessagingClient(pusherConn),
			SMSClient:        sms.NewSMSAPIClient(smsConn),
			CallClient:       call.NewCallAPIClient(callConn),
			SubscriberClient: subscriber.NewSubscriberAPIClient(subscriberConn),
		})
		errs.Panic(err)

		messaging.RegisterMessagingServer(app.GRPCServer(), messagingAPI)
		errs.Panic(messaging.RegisterMessagingHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		return nil
	})
}
