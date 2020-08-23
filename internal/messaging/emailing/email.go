package emailing

import (
	"context"

	"gopkg.in/gomail.v2"

	"github.com/gidyon/services/pkg/api/messaging/emailing"
	"github.com/gidyon/services/pkg/auth"
	"github.com/gidyon/services/pkg/utils/errs"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/grpclog"
)

type dialer interface {
	DialAndSend(...*gomail.Message) error
}

type emailingAPIServer struct {
	logger  grpclog.LoggerV2
	authAPI auth.Interface
	sender  func(*emailing.Email) error
	dialer  dialer
	opt     *Options
}

// Options contains the parameters passed while calling NewEmailingAPIServer
type Options struct {
	Logger        grpclog.LoggerV2
	JWTSigningKey []byte
	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  string
}

// NewEmailingAPIServer is singleton for creating email server APIs
func NewEmailingAPIServer(ctx context.Context, opt *Options) (emailing.EmailingServer, error) {
	// Validation
	var err error
	switch {
	case ctx == nil:
		err = errs.NilObject("context")
	case opt == nil:
		err = errs.NilObject("options")
	case opt.Logger == nil:
		err = errs.NilObject("logger")
	case opt.JWTSigningKey == nil:
		err = errs.NilObject("jwt key")
	case opt.SMTPHost == "":
		err = errs.MissingField("smtp host")
	case opt.SMTPPort == 0:
		err = errs.MissingField("smtp port")
	case opt.SMTPUsername == "":
		err = errs.MissingField("smtp username")
	case opt.SMTPPassword == "":
		err = errs.MissingField("smtp password")
	}
	if err != nil {
		return nil, err
	}

	// Auth API
	authAPI, err := auth.NewAPI(opt.JWTSigningKey, "EMailing API", "users")
	if err != nil {
		return nil, err
	}

	// SMTP dialer
	dialer := &gomail.Dialer{
		Host:     opt.SMTPHost,
		Port:     opt.SMTPPort,
		Username: opt.SMTPUsername,
		Password: opt.SMTPPassword,
	}

	// API
	emailingAPI := &emailingAPIServer{
		logger:  opt.Logger,
		authAPI: authAPI,
		opt:     opt,
		dialer:  dialer,
	}

	emailingAPI.sender = emailingAPI.sendEmail

	return emailingAPI, nil
}

func (api *emailingAPIServer) SendEmail(
	ctx context.Context, sendReq *emailing.Email,
) (*empty.Empty, error) {
	// Validate email
	err := validateEmail(sendReq)
	if err != nil {
		return nil, err
	}

	// Authenticate request
	err = api.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Send email
	go func() {
		err = api.sender(sendReq)
		if err != nil {
			api.logger.Errorf("failed to send email: %v", err)
		}
	}()

	return &empty.Empty{}, nil
}

func validateEmail(email *emailing.Email) error {
	var err error
	switch {
	case email == nil:
		err = errs.NilObject("email")
	case len(email.Destinations) == 0:
		err = errs.MissingField("destinations")
	case email.From == "":
		err = errs.MissingField("from")
	case email.Subject == "":
		err = errs.MissingField("subject")
	case email.Body == "":
		err = errs.MissingField("body")
	case email.BodyContentType == "":
		email.BodyContentType = "text/html"
	}
	return err
}
