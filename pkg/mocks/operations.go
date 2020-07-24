package mocks

import (
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/operation"
	"github.com/gidyon/services/pkg/mocks/mocks"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// OperationsClientMock is a mock for operations API client
type OperationsClientMock interface {
	operation.OperationAPIClient
}

// OperationsAPI is mocked object for creating Operations APIs
var OperationsAPI = &mocks.OperationsClientMock{}

func init() {
	OperationsAPI.On("CreateOperation", mock.Anything, mock.Anything, mock.Anything).
		Return(fakeOperation(), nil)

	OperationsAPI.On("UpdateOperation", mock.Anything, mock.Anything, mock.Anything).
		Return(fakeOperation(), nil)

	OperationsAPI.On("DeleteOperation", mock.Anything, mock.Anything, mock.Anything).
		Return(&empty.Empty{}, nil)

	OperationsAPI.On("ListOperations", mock.Anything, mock.Anything, mock.Anything).
		Return(&operation.ListOperationsResponse{
			Operations: []*operation.Operation{fakeOperation(), fakeOperation()},
		}, nil)

	OperationsAPI.On("GetOperation", mock.Anything, mock.Anything, mock.Anything).
		Return(fakeOperation(), nil)
}

func fakeOperation() *operation.Operation {
	return &operation.Operation{
		Id:           uuid.New().String(),
		UserId:       uuid.New().String(),
		Details:      randomdata.Paragraph(),
		Result:       randomdata.Paragraph(),
		Status:       operation.OperationStatus(randomdata.Number(1, len(operation.OperationStatus_name))),
		TimestampSec: time.Now().Unix(),
	}
}
