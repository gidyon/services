package main

import (
	"context"
	"os"
	"time"

	"github.com/gidyon/micro"
	"github.com/gidyon/micro/pkg/grpc/auth"
	httpmiddleware "github.com/gidyon/micro/pkg/http"
	"github.com/gidyon/micro/utils/encryption"
	"github.com/gidyon/micro/utils/errs"
	"github.com/gidyon/services/pkg/api/messaging/call"
	"github.com/gidyon/services/pkg/api/subscriber"
	"github.com/gorilla/securecookie"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/gidyon/services/pkg/api/messaging/sms"

	"github.com/gidyon/services/pkg/api/messaging/pusher"

	"github.com/gidyon/services/pkg/api/messaging/emailing"

	"github.com/gidyon/services/pkg/api/messaging"

	"github.com/gidyon/micro/pkg/healthcheck"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

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

	// Servemux option for JSON Marshaling
	app.AddServeMuxOptions(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
		},
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

		// Pagination hasher
		paginationHasher, err := encryption.NewHasher(string(jwtKey))
		errs.Panic(err)

		// Create messaging API instance
		messagingAPI, err := messaging_app.NewMessagingServer(ctx, &messaging_app.Options{
			SQLDBWrites:      app.GormDBByName("sqlWrites"),
			SQLDBReads:       app.GormDBByName("sqlReads"),
			Logger:           app.Logger(),
			EmailSender:      os.Getenv("SENDER_EMAIL_ADDRESS"),
			EmailClient:      emailing.NewEmailingClient(emailConn),
			PushClient:       pusher.NewPushMessagingClient(pusherConn),
			SMSClient:        sms.NewSMSAPIClient(smsConn),
			CallClient:       call.NewCallAPIClient(callConn),
			SubscriberClient: subscriber.NewSubscriberAPIClient(subscriberConn),
			AuthAPI:          authAPI,
			PaginationHasher: paginationHasher,
		})
		errs.Panic(err)

		messaging.RegisterMessagingServer(app.GRPCServer(), messagingAPI)
		errs.Panic(messaging.RegisterMessagingHandler(ctx, app.RuntimeMux(), app.ClientConn()))

		return nil
	})
}
