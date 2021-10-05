package sms

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/messaging/sms"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type smsAPIServer struct {
	sms.UnimplementedSMSAPIServer
	*Options
	mu               *sync.RWMutex // guards credentialsCache
	credentialsCache map[string]*sms.SenderCredential
}

// Options contains parameters passed while calling NewSMSAPIServer
type Options struct {
	Logger     grpclog.LoggerV2
	SQLDB      *gorm.DB
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
	case opt.SQLDB == nil:
		err = errs.MissingField("sql db")
	case opt.HTTPClient == nil:
		opt.HTTPClient = http.DefaultClient
	}
	if err != nil {
		return nil, err
	}

	// API server
	smsAPI := &smsAPIServer{
		Options:          opt,
		mu:               &sync.RWMutex{},
		credentialsCache: make(map[string]*sms.SenderCredential),
	}

	// Automigrate
	if !smsAPI.SQLDB.Migrator().HasTable(&SenderCredential{}) {
		err = smsAPI.SQLDB.AutoMigrate(&SenderCredential{})
		if err != nil {
			return nil, fmt.Errorf("automigration of %s table failed", (&SenderCredential{}).TableName())
		}
	}

	return smsAPI, nil
}

func (api *smsAPIServer) SendSMS(
	ctx context.Context, req *sms.SendSMSRequest,
) (*empty.Empty, error) {
	// Validation
	switch {
	case req == nil:
		return nil, errs.NilObject("send sms request")
	case req.Auth == nil && !req.FetchSender:
		return nil, errs.NilObject("sender auth data")
	case req.GetAuth().GetApiKey() == "" && !req.FetchSender:
		return nil, errs.MissingField("sender api key")
	case req.GetAuth().GetClientId() == "" && !req.FetchSender:
		return nil, errs.MissingField("sender client id")
	case req.GetAuth().GetAccessKey() == "" && !req.FetchSender:
		return nil, errs.MissingField("sender access key")
	case req.GetAuth().GetSenderId() == "" && !req.FetchSender:
		return nil, errs.MissingField("sender id")
	case req.FetchSender && req.ProjectId == "":
		return nil, errs.MissingField("project id")
	default:
		// Validate sms
		err := validateSMS(req.Sms)
		if err != nil {
			return nil, err
		}
	}

	// Get sender id locally
	if req.FetchSender {
		getRes, err := api.GetSenderCredential(ctx, &sms.GetSenderCredentialRequest{
			UseProjectId: true,
			CredentialId: req.ProjectId,
		})
		if err != nil {
			return nil, err
		}
		req.Auth = getRes.Auth
	}

	// Send sms
	switch req.GetProvider() {
	case sms.SmsProvider_ONFON:
		go api.sendSmsOnfon(ctx, req)
	}

	return &empty.Empty{}, nil
}

func (api *smsAPIServer) CreateSenderCredential(
	ctx context.Context, req *sms.CreateSenderCredentialsRequest,
) (*empty.Empty, error) {
	// Validation
	switch {
	case req == nil:
		return nil, errs.MissingField("request")
	case req.Credential == nil:
		return nil, errs.MissingField("credential")
	case req.Credential.Auth == nil:
		return nil, errs.MissingField("auth")
	case req.Credential.ProjectId == "":
		return nil, errs.MissingField("project_id")
	}

	db, err := SenderCredentialModel(req.Credential)
	if err != nil {
		return nil, err
	}

	err = api.SQLDB.Create(db).Error
	switch {
	case err == nil:
	case strings.Contains(strings.ToLower(err.Error()), "duplicate"):
		return nil, errs.DoesExist("sender project", req.Credential.ProjectId)
	default:
		return nil, errs.FailedToSave("sender credential", err)
	}

	return &emptypb.Empty{}, nil
}

func (api *smsAPIServer) GetSenderCredential(
	ctx context.Context, req *sms.GetSenderCredentialRequest,
) (*sms.SenderCredential, error) {
	// Validation
	switch {
	case req == nil:
		return nil, errs.MissingField("request")
	case req.CredentialId == "":
		return nil, errs.MissingField("credential id")
	}

	// Get from cache
	if req.UseProjectId {
		api.mu.RLock()
		pb, ok := api.credentialsCache[fmt.Sprintf("#%s", req.CredentialId)]
		if ok {
			api.mu.RUnlock()
			return pb, nil
		}
		api.mu.RUnlock()
	} else {
		api.mu.RLock()
		pb, ok := api.credentialsCache[req.CredentialId]
		if ok {
			api.mu.RUnlock()
			return pb, nil
		}
		api.mu.RUnlock()
	}

	db := &SenderCredential{}

	var err error
	if req.UseProjectId {
		err = api.SQLDB.First(db, "project_id=?", req.CredentialId).Error
	} else {
		err = api.SQLDB.First(db, "id=?", req.CredentialId).Error
	}
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.DoesNotExist("sender credential", req.CredentialId)
	default:
		return nil, errs.FailedToFind("sender credential", err)
	}

	pb, err := SenderCredentialProto(db)
	if err != nil {
		return nil, err
	}

	api.mu.Lock()
	defer api.mu.Unlock()

	// Update cache
	if req.UseProjectId {
		api.credentialsCache[fmt.Sprintf("#%s", req.CredentialId)] = pb
	} else {
		api.credentialsCache[req.CredentialId] = pb
	}

	return pb, nil
}

func validateSMS(smsPB *sms.SMS) error {
	var err error
	switch {
	case smsPB == nil:
		err = errs.MissingField("sms")
	case len(smsPB.DestinationPhones) == 0:
		err = errs.MissingField("destination phones")
	case smsPB.Keyword == "":
		err = errs.MissingField("keyword")
	case smsPB.Message == "":
		err = errs.MissingField("message")
	}
	return err
}
