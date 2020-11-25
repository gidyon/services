package longrunning

import (
	"context"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/operation"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func mockOperation() *operation.Operation {
	return &operation.Operation{
		Id:           uuid.New().String(),
		UserId:       uuid.New().String(),
		Details:      randomdata.Paragraph(),
		Result:       randomdata.Paragraph(),
		Status:       operation.OperationStatus(randomdata.Number(1, len(operation.OperationStatus_name))),
		TimestampSec: time.Now().Unix(),
	}
}

var _ = Describe("Creating A Operation @create", func() {
	var (
		createReq *operation.CreateOperationRequest
		ctx       context.Context
	)

	BeforeEach(func() {
		createReq = &operation.CreateOperationRequest{
			Operation: mockOperation(),
		}
		ctx = context.Background()
	})

	Describe("Creating an operation with incorrect/missing values", func() {
		It("should fail when the request is nil", func() {
			createReq = nil
			createRes, err := OperationAPI.CreateOperation(ctx, createReq)
			Expect(err).Should(HaveOccurred())
			Expect(createRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when operation is empty", func() {
			createReq.Operation = nil
			createRes, err := OperationAPI.CreateOperation(ctx, createReq)
			Expect(err).Should(HaveOccurred())
			Expect(createRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when operation user id missing", func() {
			createReq.Operation.UserId = ""
			createRes, err := OperationAPI.CreateOperation(ctx, createReq)
			Expect(err).Should(HaveOccurred())
			Expect(createRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when operation details missing", func() {
			createReq.Operation.Details = ""
			createRes, err := OperationAPI.CreateOperation(ctx, createReq)
			Expect(err).Should(HaveOccurred())
			Expect(createRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when operation status is missing", func() {
			createReq.Operation.Status = operation.OperationStatus_OPERATION_STATUS_UNSPECIFIED
			createRes, err := OperationAPI.CreateOperation(ctx, createReq)
			Expect(err).Should(HaveOccurred())
			Expect(createRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
	})

	Describe("Creating an operation with correct/valid request", func() {
		var operationID string
		It("should succeed when the request is valid", func() {
			createRes, err := OperationAPI.CreateOperation(ctx, createReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(createRes).ShouldNot(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			operationID = createRes.Id
		})

		Describe("Created operation", func() {
			It("should exist", func() {
				opStr, err := OperationAPIService.redisDB.Get(ctx, getOpKey(operationID)).Result()
				Expect(err).ShouldNot(HaveOccurred())
				opPB := &operation.Operation{}
				Expect(proto.Unmarshal([]byte(opStr), opPB)).ShouldNot(HaveOccurred())
				Expect(opPB.Id).ShouldNot(BeZero())
				Expect(opPB.UserId).ShouldNot(BeZero())
				Expect(opPB.Details).ShouldNot(BeZero())
				Expect(opPB.Status).ShouldNot(Equal(operation.OperationStatus_OPERATION_STATUS_UNSPECIFIED))
			})
		})
	})
})
