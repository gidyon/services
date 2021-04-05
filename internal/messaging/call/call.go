package call

import (
	"context"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/messaging/call"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/grpclog"
)

type callAPIServer struct {
	call.UnimplementedCallAPIServer
	*Options
}

// Options contains the parameters passed while calling NewCallAPIServer
type Options struct {
	Logger  grpclog.LoggerV2
	AuthAPI auth.API
}

// NewCallAPIServer creates a new call API server
func NewCallAPIServer(ctx context.Context, opt *Options) (call.CallAPIServer, error) {
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
		err = errs.NilObject("auth API")
	}
	if err != nil {
		return nil, err
	}

	// API
	callAPI := &callAPIServer{
		Options: opt,
	}

	return callAPI, nil
}

func (api *callAPIServer) Call(
	ctx context.Context, callReq *call.CallPayload,
) (*empty.Empty, error) {
	// Request must not be nil
	if callReq == nil {
		return nil, errs.NilObject("Call")
	}

	// Authenticate request
	err := api.AuthAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validate call
	err = validateCall(callReq)
	if err != nil {
		return nil, err
	}

	// Send call

	return &empty.Empty{}, nil
}

func validateCall(callPB *call.CallPayload) error {
	var err error
	switch {
	case len(callPB.DestinationPhones) == 0:
		err = errs.MissingField("destination phones")
	case callPB.Keyword == "":
		err = errs.MissingField("keyword")
	case callPB.Message == "":
		err = errs.MissingField("message")
	}
	return err
}
