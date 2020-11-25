package main

import (
	"context"
	"os"

	"github.com/gidyon/micro"
	"github.com/gorilla/securecookie"

	httpmiddleware "github.com/gidyon/micro/pkg/http"
	"github.com/gidyon/micro/utils/healthcheck"

	subscriber_app "github.com/gidyon/services/internal/subscriber"

	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/api/channel"
	"github.com/gidyon/services/pkg/api/subscriber"
	"github.com/gidyon/services/pkg/auth"
	"github.com/gidyon/services/pkg/utils/encryption"
	"github.com/gidyon/services/pkg/utils/errs"

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
	cfg, err := config.New(config.FromFile)
	errs.Panic(err)

	// Create service
	app, err := micro.NewService(ctx, cfg, nil)
	errs.Panic(err)

	// Add middlewares
	recoveryUIs, recoverySIs := app_grpc_middleware.AddRecovery()
	app.AddGRPCUnaryServerInterceptors(recoveryUIs...)
	app.AddGRPCStreamServerInterceptors(recoverySIs...)

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
		// Connect to account service
		accountCC, err := app.DialExternalService(ctx, "account")
		errs.Panic(err)
		app.Logger().Infoln("connected to account service")

		// Connect to channel service
		channelCC, err := app.DialExternalService(ctx, "channel")
		errs.Panic(err)
		app.Logger().Infoln("connected to channel service")

		app.Logger().Infoln("connected to all services")

		// Create subscriber API
		subscriberAPI, err := subscriber_app.NewSubscriberAPIServer(ctx, &subscriber_app.Options{
			SQLDB:         app.GormDBByName("sqlWrites"),
			Logger:        app.Logger(),
			ChannelClient: channel.NewChannelAPIClient(channelCC),
			AccountClient: account.NewAccountAPIClient(accountCC),
			JWTSigningKey: []byte(os.Getenv("JWT_SIGNING_KEY")),
		})
		errs.Panic(err)

		subscriber.RegisterSubscriberAPIServer(app.GRPCServer(), subscriberAPI)
		errs.Panic(subscriber.RegisterSubscriberAPIHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		return nil
	})
}
