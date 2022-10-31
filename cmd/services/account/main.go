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
	"github.com/gorilla/securecookie"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/grpc/resolver"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/gorm"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/gidyon/micro/v2"
	"github.com/gidyon/micro/v2/pkg/middleware/grpc/zaplogger"
	"github.com/gidyon/micro/v2/utils/encryption"

	"github.com/gidyon/micro/v2/pkg/healthcheck"
	"github.com/gidyon/services/internal/pkg/fauth"

	account_app "github.com/gidyon/services/internal/account"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/api/messaging"

	"github.com/gidyon/micro/v2/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/v2/pkg/middleware/grpc"
	http_middleware "github.com/gidyon/micro/v2/pkg/middleware/http"

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

	c := cors.New(cors.Options{
		AllowedOrigins:       []string{"*"},
		AllowedMethods:       []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:       []string{"Accept", "Access-Control-Allow-Origin", "Authorization", "Cache-Control", "Content-Type", "DNT", "If-Modified-Since", "Keep-Alive", "Origin", "User-Agent", "X-Requested-With"},
		ExposedHeaders:       []string{"Authorization"},
		MaxAge:               1728,
		AllowCredentials:     true,
		OptionsPassthrough:   false,
		OptionsSuccessStatus: 0,
		Debug:                true,
	})

	app.AddHTTPMiddlewares(func(h http.Handler) http.Handler {
		return c.Handler(h)
	})

	apiHashKey, err := encryption.ParseKey([]byte(os.Getenv("API_HASH_KEY")))
	errs.Panic(err)

	apiBlockKey, err := encryption.ParseKey([]byte(os.Getenv("API_BLOCK_KEY")))
	errs.Panic(err)

	sc := securecookie.New(apiHashKey, apiBlockKey)

	app.AddHTTPMiddlewares(http_middleware.CookieToJWTMiddleware(&http_middleware.CookieJWTOptions{
		SecureCookie: sc,
		AuthHeader:   auth.Header(),
		AuthScheme:   auth.Scheme(),
		CookieName:   auth.JWTCookie(),
	}))

	// Grpc Gateway options
	app.AddServeMuxOptions(
		runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
			switch key {
			case "set-cookie", "access-control-expose-headers":
				return key, true
			default:
				return runtime.DefaultHeaderMatcher(key)
			}
		}),
	)

	// 5. Bootstrapping service
	app.Start(ctx, func() error {
		// Connect to messaging service
		messagingCC, err := app.ExternalServiceConn("messaging")
		errs.Panic(err)

		var firebaseAuth fauth.FirebaseAuthClient

		// Firebase app
		credFile := os.Getenv("FIREBASE_CREDENTIALS_FILE")
		if credFile != "" {
			opt := option.WithCredentialsFile(os.Getenv("FIREBASE_CREDENTIALS_FILE"))
			firebaseApp, err := firebase.NewApp(ctx, nil, opt)
			errs.Panic(err)

			// Firebase auth
			firebaseAuth, err = firebaseApp.Auth(ctx)
			errs.Panic(err)
		}

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
			PaginationHasher:   paginationHasher,
			AuthAPI:            authAPI,
			SQLDBWrites: func() *gorm.DB {
				if os.Getenv("DB_DEBUG") != "" {
					return app.GormDBByName("sqlWrites").Debug()
				}
				return app.GormDBByName("sqlWrites")
			}(),
			SQLDBReads: func() *gorm.DB {
				if os.Getenv("DB_DEBUG") != "" {
					return app.GormDBByName("sqlReads").Debug()
				}
				return app.GormDBByName("sqlReads")
			}(),
			RedisDBWrites:   app.RedisClientByName("redisWrites"),
			RedisDBReads:    app.RedisClientByName("redisReads"),
			SecureCookie:    sc,
			Logger:          app.Logger(),
			MessagingClient: messaging.NewMessagingClient(messagingCC),
			FirebaseAuth:    firebaseAuth,
		})
		errs.Panic(err)

		account.RegisterAccountAPIServer(app.GRPCServer(), accountAPI)
		errs.Panic(account.RegisterAccountAPIHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		// Downloading users API
		app.AddEndpointFunc("/api/accounts/downloads/users", downloadUsersHandler(accountAPI, authAPI))

		return nil
	})
}
