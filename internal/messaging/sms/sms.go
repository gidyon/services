package sms

import (
	"context"
	"net/http"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/messaging/sms"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/grpclog"
)

type smsAPIServer struct {
	sms.UnimplementedSMSAPIServer
	*Options
}

// Options contains parameters passed while calling NewSMSAPIServer
type Options struct {
	Logger     grpclog.LoggerV2
	AuthAPI    auth.API
	HTTPClient *http.Client
}

// NewSMSAPIServer creates a new sms API server
func NewSMSAPIServer(ctx context.Context, opt *Options) (sms.SMSAPIServer, error) {
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
	case opt.HTTPClient == nil:
		opt.HTTPClient = http.DefaultClient
	}
	if err != nil {
		return nil, err
	}

	// API server
	smsAPI := &smsAPIServer{
		Options: opt,
	}

	return smsAPI, nil
}

func (api *smsAPIServer) SendSMS(
	ctx context.Context, sendReq *sms.SendSMSRequest,
) (*empty.Empty, error) {
	// Validation
	switch {
	case sendReq == nil:
		return nil, errs.NilObject("send sms request")
	case sendReq.Auth == nil:
		return nil, errs.NilObject("sender auth data")
	case sendReq.Auth.ApiKey == "":
		return nil, errs.MissingField("sender api key")
	case sendReq.Auth.ClientId == "":
		return nil, errs.MissingField("sender client id")
	case sendReq.Auth.AccessKey == "":
		return nil, errs.MissingField("sender access key")
	case sendReq.Auth.SenderId == "":
		return nil, errs.MissingField("sender id")
	default:
		// Validate sms
		err := validateSMS(sendReq.Sms)
		if err != nil {
			return nil, err
		}
	}

	// Send sms
	switch sendReq.GetProvider() {
	case sms.SmsProvider_ONFON:
		go api.sendSmsOnfon(ctx, sendReq)
	}

	return &empty.Empty{}, nil
}

func validateSMS(smsPB *sms.SMS) error {
	var err error
	switch {
	case len(smsPB.DestinationPhones) == 0:
		err = errs.MissingField("destination phones")
	case smsPB.Keyword == "":
		err = errs.MissingField("keyword")
	case smsPB.Message == "":
		err = errs.MissingField("message")
	}
	return err
}
