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
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/gorilla/securecookie"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/gidyon/micro"
	httpmiddleware "github.com/gidyon/micro/pkg/http"
	"github.com/gidyon/micro/utils/encryption"
	"github.com/gidyon/micro/v2/pkg/middleware/grpc/zaplogger"

	"github.com/gidyon/micro/pkg/healthcheck"

	account_app "github.com/gidyon/services/internal/account"

	"github.com/gidyon/micro/pkg/grpc/auth"
	"github.com/gidyon/micro/utils/errs"
	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/api/messaging"

	"github.com/gidyon/micro/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

func main() {
	ctx := context.Background()

	apiHashKey, err := encryption.ParseKey([]byte(os.Getenv("API_HASH_KEY")))
	errs.Panic(err)

	apiBlockKey, err := encryption.ParseKey([]byte(os.Getenv("API_BLOCK_KEY")))
	errs.Panic(err)

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

	// Fetch groups
	app.AddEndpointFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		json.NewEncoder(w).Encode(auth.Groups())
	})

	sc := securecookie.New(apiHashKey, apiBlockKey)

	// Cookie based authentication
	app.AddHTTPMiddlewares(httpmiddleware.CookieToJWTMiddleware(&httpmiddleware.CookieJWTOptions{
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
			Group:       auth.AdminGroup(),
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
		messagingCC, err := app.DialExternalService(ctx, "messaging")
		errs.Panic(err)
		app.Logger().Infoln("connected to messaging service")

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
			SecureCookie:       sc,
			MessagingClient:    messaging.NewMessagingClient(messagingCC),
			FirebaseAuth:       firebaseAuth,
		})
		errs.Panic(err)

		account.RegisterAccountAPIServer(app.GRPCServer(), accountAPI)
		errs.Panic(account.RegisterAccountAPIHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		return nil
	})
}
