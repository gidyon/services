package longrunning

import (
	"context"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/longrunning"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("Updating A Operation @update", func() {
	var (
		userID    = randomdata.RandStringRunes(32)
		updateReq *longrunning.UpdateOperationRequest
		ctx       context.Context
	)

	BeforeEach(func() {
		updateReq = &longrunning.UpdateOperationRequest{
			OperationId: randomdata.RandStringRunes(32),
			Result:      randomdata.Paragraph(),
			Status:      longrunning.OperationStatus(randomdata.Number(1, len(longrunning.OperationStatus_name))),
		}
		ctx = context.Background()
	})

	Describe("Updating a operation with incorrect/missing values", func() {
		It("should fail when the request is empty", func() {
			updateReq = nil
			updateRes, err := OperationAPI.UpdateOperation(ctx, updateReq)
			Expect(err).Should(HaveOccurred())
			Expect(updateRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when operation id is missing", func() {
			updateReq.OperationId = ""
			updateRes, err := OperationAPI.UpdateOperation(ctx, updateReq)
			Expect(err).Should(HaveOccurred())
			Expect(updateRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when operation result is missing", func() {
			updateReq.Result = ""
			updateRes, err := OperationAPI.UpdateOperation(ctx, updateReq)
			Expect(err).Should(HaveOccurred())
			Expect(updateRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when operation status is missing", func() {
			updateReq.Status = longrunning.OperationStatus_OPERATION_STATUS_UNSPECIFIED
			updateRes, err := OperationAPI.UpdateOperation(ctx, updateReq)
			Expect(err).Should(HaveOccurred())
			Expect(updateRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
	})

	Describe("Updating a operation with correct/valid request", func() {
		var operationID string
		Describe("Lets create the operation first", func() {
			It("should succeed", func() {
				createReq := &longrunning.CreateOperationRequest{
					Operation: mockOperation(),
				}
				createReq.Operation.UserId = userID
				createRes, err := OperationAPI.CreateOperation(ctx, createReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(createRes).ShouldNot(BeNil())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				operationID = createRes.Id
			})

			Describe("updating the operation", func() {
				var result string
				BeforeEach(func() {
					updateReq.OperationId = operationID
					updateReq.Result = randomdata.Paragraph()
				})

				It("should succeed when the request is valid", func() {
					updateRes, err := OperationAPI.UpdateOperation(ctx, updateReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(updateRes).ShouldNot(BeNil())
					Expect(updateReq.Result).Should(Equal(updateRes.Result))
					result = updateReq.Result
				})

				Describe("Getting an updated operation", func() {
					It("should reflect updated fields", func() {
						ops, err := OperationAPIService.RedisClient.LRange(ctx, getUserOpList(userID), 0, -1).Result()
						Expect(err).ShouldNot(HaveOccurred())
						Expect(operationID).Should(BeElementOf(ops))

						opStr, err := OperationAPIService.RedisClient.Get(ctx, getOpKey(operationID)).Result()
						Expect(err).ShouldNot(HaveOccurred())
						opPB := &longrunning.Operation{}
						Expect(proto.Unmarshal([]byte(opStr), opPB)).ShouldNot(HaveOccurred())
						Expect(opPB.Id).ShouldNot(BeZero())
						Expect(opPB.UserId).ShouldNot(BeZero())
						Expect(opPB.Details).ShouldNot(BeZero())
						Expect(opPB.Status).ShouldNot(Equal(longrunning.OperationStatus_OPERATION_STATUS_UNSPECIFIED))
						Expect(opPB.Result).Should(Equal(result))
					})
				})
			})

			Describe("Updating an operation that does not exist", func() {
				var operationID = randomdata.RandStringRunes(32)
				BeforeEach(func() {
					updateReq.OperationId = operationID
					updateReq.Result = randomdata.Paragraph()
				})

				It("should fail because the operation id is unknown", func() {
					updateRes, err := OperationAPI.UpdateOperation(ctx, updateReq)
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.NotFound))
					Expect(updateRes).Should(BeNil())
				})

				Describe("Getting an updated operation", func() {
					It("should fail because the operation never existed", func() {
						opStr, err := OperationAPIService.RedisClient.Get(ctx, getOpKey(operationID)).Result()
						Expect(err).Should(HaveOccurred())
						Expect(opStr).Should(BeZero())
					})
				})
			})
		})
	})
})
