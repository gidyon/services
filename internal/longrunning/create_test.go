package longrunning

import (
	"context"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/longrunning"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func mockOperation() *longrunning.Operation {
	return &longrunning.Operation{
		Id:           uuid.New().String(),
		UserId:       uuid.New().String(),
		Details:      randomdata.Paragraph(),
		Result:       randomdata.Paragraph(),
		Status:       longrunning.OperationStatus(randomdata.Number(1, len(longrunning.OperationStatus_name))),
		TimestampSec: time.Now().Unix(),
	}
}

var _ = Describe("Creating A Operation @create", func() {
	var (
		createReq *longrunning.CreateOperationRequest
		ctx       context.Context
	)

	BeforeEach(func() {
		createReq = &longrunning.CreateOperationRequest{
			Operation: mockOperation(),
		}
		ctx = context.Background()
	})

	Describe("Creating an longrunning with incorrect/missing values", func() {
		It("should fail when the request is nil", func() {
			createReq = nil
			createRes, err := OperationAPI.CreateOperation(ctx, createReq)
			Expect(err).Should(HaveOccurred())
			Expect(createRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when longrunning is empty", func() {
			createReq.Operation = nil
			createRes, err := OperationAPI.CreateOperation(ctx, createReq)
			Expect(err).Should(HaveOccurred())
			Expect(createRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when longrunning user id missing", func() {
			createReq.Operation.UserId = ""
			createRes, err := OperationAPI.CreateOperation(ctx, createReq)
			Expect(err).Should(HaveOccurred())
			Expect(createRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when longrunning details missing", func() {
			createReq.Operation.Details = ""
			createRes, err := OperationAPI.CreateOperation(ctx, createReq)
			Expect(err).Should(HaveOccurred())
			Expect(createRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when longrunning status is missing", func() {
			createReq.Operation.Status = longrunning.OperationStatus_OPERATION_STATUS_UNSPECIFIED
			createRes, err := OperationAPI.CreateOperation(ctx, createReq)
			Expect(err).Should(HaveOccurred())
			Expect(createRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
	})

	Describe("Creating an longrunning with correct/valid request", func() {
		var longrunningID string
		It("should succeed when the request is valid", func() {
			createRes, err := OperationAPI.CreateOperation(ctx, createReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(createRes).ShouldNot(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			longrunningID = createRes.Id
		})

		Describe("Created longrunning", func() {
			It("should exist", func() {
				opStr, err := OperationAPIService.RedisClient.Get(ctx, getOpKey(longrunningID)).Result()
				Expect(err).ShouldNot(HaveOccurred())
				opPB := &longrunning.Operation{}
				Expect(proto.Unmarshal([]byte(opStr), opPB)).ShouldNot(HaveOccurred())
				Expect(opPB.Id).ShouldNot(BeZero())
				Expect(opPB.UserId).ShouldNot(BeZero())
				Expect(opPB.Details).ShouldNot(BeZero())
				Expect(opPB.Status).ShouldNot(Equal(longrunning.OperationStatus_OPERATION_STATUS_UNSPECIFIED))
			})
		})
	})
})
