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
	"google.golang.org/api/option"

	"github.com/gorilla/securecookie"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/gidyon/micro"
	httpmiddleware "github.com/gidyon/micro/pkg/http"

	"github.com/gidyon/micro/utils/healthcheck"

	account_app "github.com/gidyon/services/internal/account"

	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/api/messaging"
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

	cfg, err := config.New(config.FromFile)
	errs.Panic(err)

	app, err := micro.NewService(ctx, cfg, nil)
	errs.Panic(err)

	// Recovery middleware
	recoveryUIs, recoverySIs := app_grpc_middleware.AddRecovery()
	app.AddGRPCUnaryServerInterceptors(recoveryUIs...)
	app.AddGRPCStreamServerInterceptors(recoverySIs...)

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

	app.AddEndpointFunc("/api/accounts/token/admin", func(w http.ResponseWriter, r *http.Request) {
		authAPI, err := auth.NewAPI([]byte(os.Getenv("JWT_SIGNING_KEY")), "me", "me")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

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

	// 5. Bootstrapping service
	app.Start(ctx, func() error {
		// Connect to messaging service
		messagingCC, err := app.DialExternalService(ctx, "messaging")
		errs.Panic(err)
		app.Logger().Infoln("connected to messaging service")

		opt := option.WithCredentialsFile(os.Getenv("FIREBASE_CREDENTIALS_FILE"))
		firebaseApp, err := firebase.NewApp(ctx, nil, opt)
		errs.Panic(err)

		firebaseAuth, err := firebaseApp.Auth(ctx)
		errs.Panic(err)

		// Create account API instance
		accountAPI, err := account_app.NewAccountAPI(ctx, &account_app.Options{
			AppName:         os.Getenv("APP_NAME"),
			TemplatesDir:    os.Getenv("TEMPLATES_DIR"),
			ActivationURL:   os.Getenv("ACTIVATION_URL"),
			JWTSigningKey:   []byte(os.Getenv("JWT_SIGNING_KEY")),
			SQLDB:           app.GormDB(),
			RedisDB:         app.RedisClient(),
			Logger:          app.Logger(),
			SecureCookie:    sc,
			MessagingClient: messaging.NewMessagingClient(messagingCC),
			FirebaseAuth:    firebaseAuth,
		})
		errs.Panic(err)

		account.RegisterAccountAPIServer(app.GRPCServer(), accountAPI)
		errs.Panic(account.RegisterAccountAPIHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		return nil
	})
}
