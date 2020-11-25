package mocks

import (
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/longrunning"
	"github.com/google/uuid"
)

// LongrunningClientMock is a mock for longrunnings API client
type LongrunningClientMock interface {
	longrunning.OperationAPIClient
}

// LongrunningAPI is mocked object for creating Longrunning APIs
// var LongrunningAPI = &mocks.LongrunningClientMock{}

// func init() {
// 	LongrunningAPI.On("CreateOperation", mock.Anything, mock.Anything, mock.Anything).
// 		Return(fakeOperation(), nil)

// 	LongrunningAPI.On("UpdateOperation", mock.Anything, mock.Anything, mock.Anything).
// 		Return(fakeOperation(), nil)

// 	LongrunningAPI.On("DeleteOperation", mock.Anything, mock.Anything, mock.Anything).
// 		Return(&empty.Empty{}, nil)

// 	LongrunningAPI.On("ListLongrunning", mock.Anything, mock.Anything, mock.Anything).
// 		Return(&longrunning.ListLongrunningResponse{
// 			Longrunning: []*longrunning.Operation{fakeOperation(), fakeOperation()},
// 		}, nil)

// 	LongrunningAPI.On("GetOperation", mock.Anything, mock.Anything, mock.Anything).
// 		Return(fakeOperation(), nil)
// }

func fakeOperation() *longrunning.Operation {
	return &longrunning.Operation{
		Id:           uuid.New().String(),
		UserId:       uuid.New().String(),
		Details:      randomdata.Paragraph(),
		Result:       randomdata.Paragraph(),
		Status:       longrunning.OperationStatus(randomdata.Number(1, len(longrunning.OperationStatus_name))),
		TimestampSec: time.Now().Unix(),
	}
}
