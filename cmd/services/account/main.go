package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	firebase "firebase.google.com/go"
	"github.com/Pallinder/go-randomdata"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/grpc/resolver"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/gidyon/micro/v2"
	"github.com/gidyon/micro/v2/pkg/middleware/grpc/zaplogger"
	"github.com/gidyon/micro/v2/utils/encryption"

	"github.com/gidyon/micro/v2/pkg/healthcheck"

	account_app "github.com/gidyon/services/internal/account"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/api/messaging"

	"github.com/gidyon/micro/v2/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/v2/pkg/middleware/grpc"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
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

	jwtKey := []byte(os.Getenv("JWT_SIGNING_KEY"))

	// Authentication API
	authAPI, err := auth.NewAPI(&auth.Options{
		SigningKey: jwtKey,
		Issuer:     "Accounts API",
		Audience:   "accounts",
	})
	errs.Panic(err)

	// Generate jwt token
	token, err := authAPI.GenToken(context.Background(), &auth.Payload{}, time.Now().Add(100*24*time.Hour))
	if err == nil {
		app.Logger().Infof("test jwt is %s", token)
	}

	app.AddGRPCUnaryServerInterceptors(grpc_auth.UnaryServerInterceptor(authAPI.AuthorizeFunc))
	app.AddGRPCStreamServerInterceptors(grpc_auth.StreamServerInterceptor(authAPI.AuthorizeFunc))

	// Readiness health check
	app.AddEndpoint("/api/accounts/health/ready", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeReadiness,
		AutoMigrator: func() error { return nil },
	}))

	// Liveness health check
	app.AddEndpoint("/api/accounts/health/live", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeLiveNess,
		AutoMigrator: func() error { return nil },
	}))

	// Grpc Gateway options
	app.AddServeMuxOptions(
		runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
			switch key {
			case "set-cookie", "access-control-expose-headers":
				return key, true
			default:
				return "", false
			}
		}),
	)

	// Default token
	app.AddEndpointFunc("/api/accounts/action/get-default-jwt", func(w http.ResponseWriter, r *http.Request) {
		token, err := authAPI.GenToken(ctx, &auth.Payload{}, time.Now().Add(time.Hour*24*365))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json")

		err = json.NewEncoder(w).Encode(map[string]string{"default_token": token})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	app.AddEndpointFunc("/api/accounts/token/admin", func(w http.ResponseWriter, r *http.Request) {

		token, err := authAPI.GenToken(r.Context(), &auth.Payload{
			ID:          fmt.Sprint(1),
			Names:       randomdata.SillyName(),
			PhoneNumber: randomdata.PhoneNumber(),
		}, time.Now().Add(6*time.Hour))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json")

		err = json.NewEncoder(w).Encode(map[string]string{"token": token})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Servemux option for JSON Marshaling
	app.AddServeMuxOptions(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
		},
	}))

	// 5. Bootstrapping service
	app.Start(ctx, func() error {
		// Connect to messaging service
		messagingCC, err := app.ExternalServiceConn("messaging")
		errs.Panic(err)

		// Firebase app
		opt := option.WithCredentialsFile(os.Getenv("FIREBASE_CREDENTIALS_FILE"))
		firebaseApp, err := firebase.NewApp(ctx, nil, opt)
		errs.Panic(err)

		// Firebase auth
		firebaseAuth, err := firebaseApp.Auth(ctx)
		errs.Panic(err)

		// Pagination hasher
		paginationHasher, err := encryption.NewHasher(string(jwtKey))
		errs.Panic(err)

		// Create account API instance
		accountAPI, err := account_app.NewAccountAPI(ctx, &account_app.Options{
			AppName:            os.Getenv("APP_NAME"),
			EmailDisplayName:   os.Getenv("EMAIL_DISPLAY_NAME"),
			DefaultEmailSender: os.Getenv("DEFAULT_EMAIL_SENDER"),
			TemplatesDir:       os.Getenv("TEMPLATES_DIR"),
			ActivationURL:      os.Getenv("ACTIVATION_URL"),
			AuthAPI:            authAPI,
			PaginationHasher:   paginationHasher,
			SQLDBWrites:        app.GormDBByName("sqlWrites"),
			SQLDBReads:         app.GormDBByName("sqlReads"),
			RedisDBWrites:      app.RedisClientByName("redisWrites"),
			RedisDBReads:       app.RedisClientByName("redisReads"),
			Logger:             app.Logger(),
			MessagingClient:    messaging.NewMessagingClient(messagingCC),
			FirebaseAuth:       firebaseAuth,
		})
		errs.Panic(err)

		account.RegisterAccountAPIServer(app.GRPCServer(), accountAPI)
		errs.Panic(account.RegisterAccountAPIHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		return nil
	})
}
