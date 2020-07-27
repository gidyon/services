package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"github.com/Pallinder/go-randomdata"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"github.com/gorilla/securecookie"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/gidyon/micro"
	httpmiddleware "github.com/gidyon/micro/pkg/http"

	"github.com/gidyon/micro/utils/healthcheck"

	account_app "github.com/gidyon/services/internal/account"

	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/api/messaging"
	"github.com/gidyon/services/pkg/auth"

	"github.com/gidyon/micro/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"
)

func main() {
	ctx := context.Background()

	apiHashKey, err := parseKeySize([]byte(os.Getenv("API_HASH_KEY")))
	handleErr(err)

	apiBlockKey, err := parseKeySize([]byte(os.Getenv("API_BLOCK_KEY")))
	handleErr(err)

	cfg, err := config.New(config.FromFile)
	handleErr(err)

	app, err := micro.NewService(ctx, cfg, nil)
	handleErr(err)

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
		CookieName:   auth.CookieName(),
	}))

	// Grpc Gateway options
	app.AddServeMuxOptions(
		runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
			switch key {
			case "Set-Cookie", "set-cookie":
				return key, true
			default:
				return "", false
			}
		}),
	)

	app.AddEndpointFunc("/api/accounts/token/admin", func(w http.ResponseWriter, r *http.Request) {
		authAPI, err := auth.NewAPI([]byte(os.Getenv("JWT_SIGNING_KEY")))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		token, err := authAPI.GenToken(r.Context(), &auth.Payload{
			ID:          fmt.Sprint(1),
			Names:       randomdata.SillyName(),
			PhoneNumber: randomdata.PhoneNumber(),
		}, 0)
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
		messagingCC, err := app.DialExternalService(ctx, "messaging", grpc.WithBlock())
		handleErr(err)
		app.Logger().Infoln("connected to messaging service")

		opt := option.WithCredentialsFile(os.Getenv("FIREBASE_CREDENTIALS_FILE"))
		firebaseApp, err := firebase.NewApp(ctx, nil, opt)
		handleErr(err)

		firebaseAuth, err := firebaseApp.Auth(ctx)
		handleErr(err)

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
		handleErr(err)

		account.RegisterAccountAPIServer(app.GRPCServer(), accountAPI)
		handleErr(account.RegisterAccountAPIHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		return nil
	})
}

func parseKeySize(key []byte) ([]byte, error) {
	keyLen := len(key)
	switch {
	case keyLen < 16:
		return nil, errors.New("key length less that 16")
	case keyLen < 24:
		return key[:16], nil
	case keyLen < 32:
		return key[:24], nil
	default:
		return key[:32], nil
	}
}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
