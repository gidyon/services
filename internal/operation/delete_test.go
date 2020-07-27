package operation

import (
	"context"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/operation"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Deleting an operation with incorrect user id
// Deleting an operation with incorrect operation id
// Deleting an operation with both operation id and user id incorrect

var _ = Describe("Deleting A Operation @delete", func() {
	var (
		userID    = randomdata.RandStringRunes(32)
		deleteReq *operation.DeleteOperationRequest
		ctx       context.Context
	)

	BeforeEach(func() {
		deleteReq = &operation.DeleteOperationRequest{
			OperationId: randomdata.RandStringRunes(32),
			UserId:      userID,
		}
		ctx = context.Background()
	})

	Describe("Deleting a operation with incorrect/missing values", func() {
		It("should fail when the request is nil", func() {
			deleteReq = nil
			deleteRes, err := OperationAPIService.DeleteOperation(ctx, deleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(deleteRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when operation id is missing", func() {
			deleteReq.OperationId = ""
			deleteRes, err := OperationAPIService.DeleteOperation(ctx, deleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(deleteRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when user id is missing", func() {
			deleteReq.UserId = ""
			deleteRes, err := OperationAPIService.DeleteOperation(ctx, deleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(deleteRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
	})

	Describe("Deleting a operation with correct/valid request", func() {
		var operationID string
		Describe("Lets create the operation first", func() {
			Describe("Creating multiple operation", func() {
				for i := 0; i < 10; i++ {
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
				}
			})
			Describe("Deleting the operation", func() {
				It("should succeed when the request is valid", func() {
					deleteReq.OperationId = operationID
					deleteReq.UserId = userID
					deleteRes, err := OperationAPIService.DeleteOperation(ctx, deleteReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(deleteRes).ShouldNot(BeNil())
					Expect(status.Code(err)).Should(Equal(codes.OK))
				})
			})

			Describe("Checking whether the operation was deleted", func() {
				It("should not exist in cache", func() {
					opStr, err := OperationAPIService.redisDB.Get(getOpKey(operationID)).Result()
					Expect(opStr).Should(BeZero())
					Expect(err).Should(HaveOccurred())

					ops, err := OperationAPIService.redisDB.LRange(getUserOpList(userID), 0, -1).Result()
					Expect(err).ShouldNot(HaveOccurred())

					Expect(operationID).ShouldNot(BeElementOf(ops))
				})
			})
		})

		Describe("Deleting an operation with incorrect user id", func() {
			It("should succeed", func() {
				deleteReq.UserId = randomdata.RandStringRunes(32)
				deleteRes, err := OperationAPIService.DeleteOperation(ctx, deleteReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(deleteRes).ShouldNot(BeNil())
				Expect(status.Code(err)).Should(Equal(codes.OK))
			})
		})

		Describe("Deleting an operation with incorrect operation id", func() {
			It("should succeed", func() {
				deleteReq.OperationId = randomdata.RandStringRunes(32)
				deleteRes, err := OperationAPIService.DeleteOperation(ctx, deleteReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(deleteRes).ShouldNot(BeNil())
				Expect(status.Code(err)).Should(Equal(codes.OK))
			})
		})

		Describe("Deleting an operation with incorrect operation idand user id", func() {
			It("should succeed", func() {
				deleteReq.OperationId = randomdata.RandStringRunes(32)
				deleteReq.UserId = randomdata.RandStringRunes(32)
				deleteRes, err := OperationAPIService.DeleteOperation(ctx, deleteReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(deleteRes).ShouldNot(BeNil())
				Expect(status.Code(err)).Should(Equal(codes.OK))
			})
		})
	})
})
