package errs

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FromJSONMarshal wraps error returned from json.Marshal to a status error
func FromJSONMarshal(err error, obj string) error {
	return status.Errorf(codes.Internal, "failed to json marshal %s: %v", obj, err)
}

// FromJSONUnMarshal wraps error returned from json.Unmarshal to a status error
func FromJSONUnMarshal(err error, obj string) error {
	return status.Errorf(codes.Internal, "failed to json unmarshal %s: %v", obj, err)
}

// FromProtoMarshal wraps error returned from proto.Marshal to a status error
func FromProtoMarshal(err error, obj string) error {
	return status.Errorf(codes.Internal, "failed to proto marshal %s: %v", obj, err)
}

// FromProtoUnMarshal wraps error returned from proto.Unmarshal to a status error
func FromProtoUnMarshal(err error, obj string) error {
	return status.Errorf(codes.Internal, "failed to proto unmarshal %s: %v", obj, err)
}

// MissingField returns a status error caused by a missing message field
func MissingField(field string) error {
	return status.Errorf(codes.InvalidArgument, "missing message field: %v", field)
}

// BadCredential creates a status error caused by bad credentials
func BadCredential(credential string) error {
	return status.Errorf(codes.InvalidArgument, "bad credential: %v", credential)
}

// CheckingCreds wraps error returned while checking credentials to a status error
func CheckingCreds(err error) error {
	return status.Errorf(codes.Internal, "failed while checking credentials: %v", err)
}

// PermissionDenied result from performing non-priviledged operations
func PermissionDenied(operation string) error {
	return status.Errorf(codes.PermissionDenied, "not authorised to perform %s operation", operation)
}

// FailedToGenToken wraps error caused while generating jwt to a status error
func FailedToGenToken(err error) error {
	return status.Errorf(codes.Internal, "failed to generate jwt: %v", err)
}

// FailedToParseToken wraps error caused while parsing token to a status error
func FailedToParseToken(err error) error {
	return status.Errorf(codes.Internal, "failed to parse jwt: %v", err)
}

// NilObject is error resulting from using nil references to objects
func NilObject(object string) error {
	return status.Errorf(codes.InvalidArgument, "nil object not allowed: %s", object)
}

// AuthenticationFailed wraps error returned from failed authentication to a proper status error
func AuthenticationFailed(err error, operation string) error {
	return status.Errorf(status.Code(err), "failed to perform operation %s: %v", operation, err)
}

// ConvertingType wraps error that occured during type assertion to grpc status error
func ConvertingType(err error, from, to string) error {
	return status.Errorf(codes.Internal, "couldn't convert from %s to %s: %v", from, to, err)
}

// IncorrectVal returns a status error indicating val was incorrect
func IncorrectVal(val string) error {
	return status.Errorf(codes.InvalidArgument, "incorrect value for %q", val)
}

// WriteFailed returns a status error for a write op error
func WriteFailed(err error) error {
	return status.Errorf(codes.Internal, "write operation failed: %v", err)
}

// ReadFailed returns a status error for a read op error
func ReadFailed(err error) error {
	return status.Errorf(codes.Internal, "read operation failed: %v", err)
}

// ActorUknown returns a status error indicating that the acto must be known
func ActorUknown() error {
	return status.Error(codes.InvalidArgument, "actor must be known")
}

// ActorNotAllowed returns a status error indicating that the acto must be known
func ActorNotAllowed() error {
	return status.Error(codes.InvalidArgument, "actor is not allowed")
}

// OperationUknown returns a status error indicating that operation must be known
func OperationUknown() error {
	return status.Error(codes.InvalidArgument, "operation must be known")
}

// ContractNotRegistered returns a status error when unregistered smart contracts access ledger
func ContractNotRegistered(contractID string) error {
	return status.Errorf(codes.PermissionDenied, "smart contract %s unknown or unregistered", contractID)
}

// CreateLogFailed returns a status error that happens during the creation of a new log
func CreateLogFailed(err error) error {
	return status.Errorf(codes.Internal, "failed to create log: %v", err)
}

// LogCastingFailed returns a status error that happens during broadcasting of new log
func LogCastingFailed(err error) error {
	return status.Errorf(codes.DataLoss, "failed to broadcast new log: %v", err)
}

// AddingLogFailed happens when adding a new log fails
func AddingLogFailed(err error) error {
	return status.Errorf(codes.Internal, "failed to add new log to ledger: %v", err)
}

// AccountDoesntExist indicates that the account does not exist
func AccountDoesntExist(accountID string) error {
	return status.Errorf(codes.NotFound, "account with id %s does not exist", accountID)
}

// AccountLoged indicates that the account has been loged
func AccountLoged() error {
	return status.Error(codes.PermissionDenied, "account has been loged - contact system admin for help")
}

// AccountAlreadyLoged indicates that the account has been loged
func AccountAlreadyLoged() error {
	return status.Error(codes.PermissionDenied, "account has already been loged - contact system admin for help")
}

// AccountNotLoged indicates that the account is not loged for it to be unblocekd
func AccountNotLoged() error {
	return status.Error(codes.PermissionDenied, "account state is not BLOCKED")
}

// AccountNotActive indicates that the account is not inactive for it to be activated
func AccountNotActive() error {
	return status.Error(codes.PermissionDenied, "account state is not ACTIVE")
}

// AccountNotDeleted indicates that the account is not inactive for it to be activated
func AccountNotDeleted() error {
	return status.Error(codes.PermissionDenied, "account state is not deleted")
}

// AccountNotInactive indicates that the account is not inactive for it to be activated
func AccountNotInactive() error {
	return status.Error(codes.PermissionDenied, "account state is not INACTIVE")
}

// AccountNotGroupMember indicates that the account is not inactive for it to be activated
func AccountNotGroupMember(group string) error {
	return status.Errorf(codes.PermissionDenied, "Group %s not associated with the account", group)
}

// OnlyOwnerPermitted indicates that only owner account is permitted to make changes
func OnlyOwnerPermitted() error {
	return status.Error(codes.PermissionDenied, "the account must be OWNER")
}

// SQLQueryFailed wraps sql error to a status error
func SQLQueryFailed(err error, queryType string) error {
	return status.Errorf(codes.Internal, "failed to execute %s query: %v", queryType, err)
}

// SQLQueryNoRows wraps sql no rows found error to a status error
func SQLQueryNoRows(err error) error {
	return status.Errorf(codes.NotFound, "no rows found for query: %v", err)
}

// FailedToSave is status erro returned from failed save operation
func FailedToSave(err error) error {
	return status.Errorf(codes.Internal, "failed to save to database: %v", err)
}

// AccountDoesExist indicates that the account already exist
func AccountDoesExist(exists, val string) error {
	return status.Errorf(codes.ResourceExhausted, "account exists with %s %s", exists, val)
}

// WrongPassword is error returned when password is incorrect
func WrongPassword() error {
	return status.Error(codes.Unauthenticated, "wrong password")
}

// FailedToGenHashedPass wraps error while generating password hash to a grpc error
func FailedToGenHashedPass(err error) error {
	return status.Errorf(codes.InvalidArgument, "failed to generate hashed password: %v", err)
}

// TokenCredentialNotMatching creates a status error caused by mismatch in token credential
func TokenCredentialNotMatching(cred string) error {
	return status.Errorf(codes.PermissionDenied, "token credential %v do not match", cred)
}

// PasswordNoMatch is error returned due to mismatch password
func PasswordNoMatch() error {
	return status.Error(codes.InvalidArgument, "passwords do not match")
}

// FailedToBeginTx wraps error returned from failed transaction to status error
func FailedToBeginTx(err error) error {
	return status.Errorf(codes.Internal, "failed to begin db transaction: %v", err)
}

// FailedToCommitTx wraps error returned from failed commit of transaction to status error
func FailedToCommitTx(err error) error {
	return status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
}

// FailedToRollbackTx wraps error returned from failed rollback of transaction to status error
func FailedToRollbackTx(err error) error {
	return status.Errorf(codes.Internal, "failed to rollback transaction: %v", err)
}

// FailedToEncrypt is status error from failed encryption
func FailedToEncrypt(err error) error {
	return status.Errorf(codes.Internal, "failed to encrypt data: %v", err)
}

// FailedToDecrypt is status error from failed decryption
func FailedToDecrypt(err error) error {
	return status.Errorf(codes.Internal, "failed to decrypt data: %v", err)
}

// FailedToPublish returns a status error when publishing to a channel fails
func FailedToPublish(err error) error {
	return status.Errorf(codes.Internal, "failed to publish message: %v", err)
}

// FailedToGetPeersNum returns a status error when getting peers number in a network
func FailedToGetPeersNum(err error) error {
	return status.Errorf(codes.Internal, "failed to get peers number: %v", err)
}

// FailedToGetLog returns a status error of error found when getting log
func FailedToGetLog(err error) error {
	return status.Errorf(codes.NotFound, "failed to get log: %v", err)
}

// FailedToGetLogs returns a status error of error found when getting log
func FailedToGetLogs(err error) error {
	return status.Errorf(codes.NotFound, "failed to get logs: %v", err)
}

// FailedToAddToledger returns a status error when adding log to ledger fails
func FailedToAddToledger(err error) error {
	return status.Errorf(status.Code(err), "failed to add log to ledger: %v", err)
}

// FailedToAddToOrderer returns a status error when adding log to orderer fails
func FailedToAddToOrderer(err error) error {
	return status.Errorf(status.Code(err), "failed to add log to orderer: %v", err)
}

// FailedToGetledger ...
func FailedToGetledger(err error) error {
	return status.Errorf(codes.Unknown, "failed to get ledger database: %v", err)
}

// LogNotFound returns a status error when a log has not been found
func LogNotFound(hash string) error {
	return status.Errorf(codes.NotFound, "log with hash: %s not found", hash)
}

// LogsNotFound returns a status error when a log has not been found
func LogsNotFound() error {
	return status.Error(codes.NotFound, "logs not found")
}

// NoMedRecordFoundForPatient returns a status when no medical record is found for a patient
func NoMedRecordFoundForPatient() error {
	return status.Error(codes.NotFound, "no medical record found for patient")
}

// HospitalNotFound is status error idicating that hospital was not found
func HospitalNotFound(hospitalID string) error {
	return status.Errorf(codes.NotFound, "hospital with id %s not found", hospitalID)
}

// HospitalsNotFound returns a status error for hospitals not found
func HospitalsNotFound() error {
	return status.Error(codes.NotFound, "no hospitals found")
}

// InsuranceNotFound is status error idicating that insurance was not found
func InsuranceNotFound(insuranceID string) error {
	return status.Errorf(codes.NotFound, "insurance with id %s not found", insuranceID)
}

// InsurancesNotFound returns a status error for insurances not found
func InsurancesNotFound() error {
	return status.Error(codes.NotFound, "no insurances found")
}

// SuperAdminLoged indicates that the super admin account has been loged
func SuperAdminLoged() error {
	return status.Error(codes.PermissionDenied, "the super admin account has been loged")
}

// SuperAdminNotActive indicates that the super admin account is not actibe
func SuperAdminNotActive() error {
	return status.Error(codes.PermissionDenied, "the super admin account state is not active")
}

// ChannelDoesntExist a status error that indicates that the channel does not exist
func ChannelDoesntExist() error {
	return status.Error(codes.NotFound, "channel does not exist")
}

// ChannelExists is a status error that indicates the channel exists
func ChannelExists(channel string) error {
	return status.Errorf(codes.AlreadyExists, "channel %s already exists", channel)
}

// SubscriberDoesntExist returns error indicating that the subscriber with given id doesn't exist
func SubscriberDoesntExist(accountID string) error {
	return status.Errorf(codes.NotFound, "subscriber account with id: %s does not exist", accountID)
}

// NotificationAccountDoesntExist returns error indicating that the notification account doesn't exist
func NotificationAccountDoesntExist(accountID string) error {
	return status.Errorf(codes.NotFound, "notification account with id: %s does not exist", accountID)
}

// RedisCmdFailed wraps error returned when redis query fails to a status error
func RedisCmdFailed(err error, queryType string) error {
	return status.Errorf(codes.Internal, "failed to execute %s command: %v", queryType, err)
}

// EmploymentNotFound returns a status error when an employment data is not found
func EmploymentNotFound(employmentID string) error {
	return status.Errorf(codes.NotFound, "employment with id %s not found", employmentID)
}

// UserEmploymentNotFound returns a status error when an employment data is not found
func UserEmploymentNotFound(employmentID, accountID string) error {
	return status.Errorf(
		codes.NotFound, "employment for user %s with id %s not found", accountID, employmentID,
	)
}

// UserEmploymentsNotFound returns a status error when an employment data is not found
func UserEmploymentsNotFound(accountID string) error {
	return status.Errorf(codes.NotFound, "employments for user %s not found", accountID)
}

// FailedToExecuteTemplate returns a status error for a failed template execution
func FailedToExecuteTemplate(err error) error {
	return status.Errorf(codes.Internal, "failed to execute template: %v", err)
}

// FailedToGetToken wraps the error returned while getting token to a status error
func FailedToGetToken(err error) error {
	return status.Errorf(codes.Internal, "failed to get token: %v", err)
}

// WrapErrorWithCode is a wraps generic error to status error with provided code
func WrapErrorWithCode(code codes.Code, err error) error {
	return status.Error(code, err.Error())
}

// WrapError is a wraps generic error to status error
func WrapError(err error) error {
	return status.Error(status.Code(err), err.Error())
}

// WrapErrorWithCodeAndMsg wraps generic error to status error with provided code and msg
func WrapErrorWithCodeAndMsg(code codes.Code, err error, msg string) error {
	return status.Errorf(code, "%s: %v", msg, err.Error())
}

// WrapErrorWithCodeAndMsgFunc is a common message wrapper for WrapErrorWithCodeAndMsg
func WrapErrorWithCodeAndMsgFunc(msg string) func(codes.Code, error) error {
	return func(code codes.Code, err error) error {
		if err != nil {
			return WrapErrorWithCodeAndMsg(code, err, msg)
		}
		return nil
	}
}

// WrapErrorWithMsg is a wraps generic error to status error with code and msg formt
func WrapErrorWithMsg(err error, msg string) error {
	return status.Errorf(status.Code(err), "%s: %v", msg, err.Error())
}

// WrapErrorWithMsgFunc is a common message wrapper for WrapErrorWithMsg
func WrapErrorWithMsgFunc(msg string) func(error) error {
	return func(err error) error {
		if err != nil {
			return WrapErrorWithMsg(err, msg)
		}
		return nil
	}
}

// WrapMessage is a wraps message provided to status error
func WrapMessage(code codes.Code, msg string) error {
	return status.Error(code, msg)
}

// WrapMessagef is a wraps message provided to status error
func WrapMessagef(code codes.Code, format string, args ...interface{}) error {
	return status.Error(code, fmt.Sprintf(format, args...))
}
