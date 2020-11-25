package longrunning

import (
	"context"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/operation"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Getting A Operation @get", func() {
	var (
		userID = randomdata.RandStringRunes(32)
		getReq *operation.GetOperationRequest
		ctx    context.Context
	)

	BeforeEach(func() {
		getReq = &operation.GetOperationRequest{
			OperationId: randomdata.RandStringRunes(32),
		}
		ctx = context.Background()
	})

	Describe("Getting an operation with incorrect/missing values", func() {
		It("should fail when the request is nil", func() {
			getReq = nil
			getRes, err := OperationAPIService.GetOperation(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(getRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when operation id is missing", func() {
			getReq.OperationId = ""
			getRes, err := OperationAPIService.GetOperation(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(getRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when operation id is incorrect", func() {
			getRes, err := OperationAPIService.GetOperation(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(getRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.NotFound))
		})
	})

	Describe("Getting a operation with correct/valid request", func() {
		var operationID string
		Describe("Lets create the operation first", func() {
			Describe("Creating an operation", func() {
				It("should succeed", func() {
					createReq := &operation.CreateOperationRequest{
						Operation: mockOperation(),
					}
					createReq.Operation.UserId = userID
					createRes, err := OperationAPIService.CreateOperation(ctx, createReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(createRes).ShouldNot(BeNil())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					operationID = createRes.Id
				})
			})

			Describe("Getting the operation", func() {
				It("should succeed when the request is valid", func() {
					getReq.OperationId = operationID
					getRes, err := OperationAPIService.GetOperation(ctx, getReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(getRes).ShouldNot(BeNil())
					Expect(status.Code(err)).Should(Equal(codes.OK))
				})
			})

			Describe("Checking whether the operation exists in cache", func() {
				It("should exist in cache", func() {
					opStr, err := OperationAPIService.redisDB.Get(ctx, getOpKey(operationID)).Result()
					Expect(opStr).ShouldNot(BeZero())
					Expect(err).ShouldNot(HaveOccurred())
				})
			})
		})
	})
})
