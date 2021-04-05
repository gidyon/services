package emailing

import (
	"context"

	"gopkg.in/gomail.v2"

	"github.com/gidyon/services/pkg/api/messaging/emailing"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/grpclog"
)

type dialer interface {
	DialAndSend(...*gomail.Message) error
}

type emailingAPIServer struct {
	emailing.UnimplementedEmailingServer
	sender func(*emailing.SendEmailRequest)
	dialer dialer
	*Options
}

// Options contains the parameters passed while calling NewEmailingAPIServer
type Options struct {
	Logger       grpclog.LoggerV2
	AuthAPI      auth.API
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
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
	case opt.AuthAPI == nil:
		err = errs.NilObject("auth api")
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

	// SMTP dialer
	dialer := &gomail.Dialer{
		Host:     opt.SMTPHost,
		Port:     opt.SMTPPort,
		Username: opt.SMTPUsername,
		Password: opt.SMTPPassword,
	}

	// API
	emailingAPI := &emailingAPIServer{
		Options: opt,
		dialer:  dialer,
	}

	emailingAPI.sender = emailingAPI.sendEmail

	return emailingAPI, nil
}

func (api *emailingAPIServer) SendEmail(
	ctx context.Context, sendReq *emailing.SendEmailRequest,
) (*empty.Empty, error) {
	// Validate email
	switch {
	case sendReq == nil:
		return nil, errs.NilObject("send request")
	case sendReq.Email == nil:
		return nil, errs.NilObject("email")
	default:
		err := validateEmail(sendReq.Email)
		if err != nil {
			return nil, err
		}
	}

	// Send email
	go api.sender(sendReq)

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
