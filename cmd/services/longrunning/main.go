package main

import (
	"context"
	"os"
	"time"

	"github.com/gidyon/micro"
	"github.com/gidyon/micro/pkg/healthcheck"
	httpmiddleware "github.com/gidyon/micro/pkg/http"
	"github.com/gorilla/securecookie"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"

	longrunning_app "github.com/gidyon/services/internal/longrunning"

	"github.com/gidyon/micro/pkg/grpc/auth"
	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"
	"github.com/gidyon/micro/utils/encryption"
	"github.com/gidyon/micro/utils/errs"
	"github.com/gidyon/services/pkg/api/longrunning"

	"github.com/gidyon/micro/pkg/config"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
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

	jwtKey := []byte(os.Getenv("JWT_SIGNING_KEY"))

	// Authentication API
	authAPI, err := auth.NewAPI(jwtKey, "USSD Log API", "users")
	errs.Panic(err)

	// Generate jwt token
	token, err := authAPI.GenToken(context.Background(), &auth.Payload{Group: auth.AdminGroup()}, time.Now().Add(time.Hour*24))
	if err == nil {
		app.Logger().Infof("Test jwt is %s", token)
	}

	app.AddGRPCUnaryServerInterceptors(grpc_auth.UnaryServerInterceptor(authAPI.AuthFunc))
	app.AddGRPCStreamServerInterceptors(grpc_auth.StreamServerInterceptor(authAPI.AuthFunc))

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

	// Servemux option for JSON Marshaling
	app.AddServeMuxOptions(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
		},
	}))

	app.Start(ctx, func() error {
		// Pagination hasher
		paginationHasher, err := encryption.NewHasher(string(jwtKey))
		errs.Panic(err)

		longrunningAPI, err := longrunning_app.NewOperationAPIService(ctx, &longrunning_app.Options{
			RedisClient:      app.RedisClient(),
			Logger:           app.Logger(),
			AuthAPI:          authAPI,
			PaginationHasher: paginationHasher,
		})
		errs.Panic(err)

		longrunning.RegisterOperationAPIServer(app.GRPCServer(), longrunningAPI)
		longrunning.RegisterOperationAPIHandler(ctx, app.RuntimeMux(), app.ClientConn())

		return nil
	})
}
