package main

import (
	"context"
	"os"

	"github.com/gidyon/micro"
	httpmiddleware "github.com/gidyon/micro/pkg/http"
	"github.com/gidyon/micro/utils/healthcheck"
	"github.com/gorilla/securecookie"

	operation_app "github.com/gidyon/services/internal/operation"

	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"
	"github.com/gidyon/services/pkg/api/operation"
	"github.com/gidyon/services/pkg/auth"
	"github.com/gidyon/services/pkg/utils/encryption"
	"github.com/gidyon/services/pkg/utils/errs"

	"github.com/gidyon/micro/pkg/config"
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
	app.AddEndpoint("/api/longrunning/health/ready", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeReadiness,
		AutoMigrator: func() error { return nil },
	}))

	// Liveness health check
	app.AddEndpoint("/api/longrunning/health/live", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
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

	app.Start(ctx, func() error {
		operationAPI, err := operation_app.NewOperationAPIService(ctx, &operation_app.Options{
			RedisClient:   app.RedisClient(),
			Logger:        app.Logger(),
			JWTSigningKey: []byte(os.Getenv("JWT_SIGNING_KEY")),
		})
		errs.Panic(err)

		operation.RegisterOperationAPIServer(app.GRPCServer(), operationAPI)
		operation.RegisterOperationAPIHandler(ctx, app.RuntimeMux(), app.ClientConn())

		return nil
	})
}
