package longrunning

import (
	"context"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/longrunning"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Deleting an longrunning with incorrect user id
// Deleting an longrunning with incorrect longrunning id
// Deleting an longrunning with both longrunning id and user id incorrect

var _ = Describe("Deleting A Operation @delete", func() {
	var (
		userID    = randomdata.RandStringRunes(32)
		deleteReq *longrunning.DeleteOperationRequest
		ctx       context.Context
	)

	BeforeEach(func() {
		deleteReq = &longrunning.DeleteOperationRequest{
			OperationId: randomdata.RandStringRunes(32),
			UserId:      userID,
		}
		ctx = context.Background()
	})

	Describe("Deleting a longrunning with incorrect/missing values", func() {
		It("should fail when the request is nil", func() {
			deleteReq = nil
			deleteRes, err := OperationAPIService.DeleteOperation(ctx, deleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(deleteRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when longrunning id is missing", func() {
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

	Describe("Deleting a longrunning with correct/valid request", func() {
		var longrunningID string
		Describe("Lets create the longrunning first", func() {
			Describe("Creating multiple longrunning", func() {
				for i := 0; i < 10; i++ {
					It("should succeed", func() {
						createReq := &longrunning.CreateOperationRequest{
							Operation: mockOperation(),
						}
						createReq.Operation.UserId = userID
						createRes, err := OperationAPIService.CreateOperation(ctx, createReq)
						Expect(err).ShouldNot(HaveOccurred())
						Expect(createRes).ShouldNot(BeNil())
						Expect(status.Code(err)).Should(Equal(codes.OK))
						longrunningID = createRes.Id
					})
				}
			})
			Describe("Deleting the longrunning", func() {
				It("should succeed when the request is valid", func() {
					deleteReq.OperationId = longrunningID
					deleteReq.UserId = userID
					deleteRes, err := OperationAPIService.DeleteOperation(ctx, deleteReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(deleteRes).ShouldNot(BeNil())
					Expect(status.Code(err)).Should(Equal(codes.OK))
				})
			})

			Describe("Checking whether the longrunning was deleted", func() {
				It("should not exist in cache", func() {
					opStr, err := OperationAPIService.RedisClient.Get(ctx, getOpKey(longrunningID)).Result()
					Expect(opStr).Should(BeZero())
					Expect(err).Should(HaveOccurred())

					ops, err := OperationAPIService.RedisClient.LRange(ctx, getUserOpList(userID), 0, -1).Result()
					Expect(err).ShouldNot(HaveOccurred())

					Expect(longrunningID).ShouldNot(BeElementOf(ops))
				})
			})
		})

		Describe("Deleting an longrunning with incorrect user id", func() {
			It("should succeed", func() {
				deleteReq.UserId = randomdata.RandStringRunes(32)
				deleteRes, err := OperationAPIService.DeleteOperation(ctx, deleteReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(deleteRes).ShouldNot(BeNil())
				Expect(status.Code(err)).Should(Equal(codes.OK))
			})
		})

		Describe("Deleting an longrunning with incorrect longrunning id", func() {
			It("should succeed", func() {
				deleteReq.OperationId = randomdata.RandStringRunes(32)
				deleteRes, err := OperationAPIService.DeleteOperation(ctx, deleteReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(deleteRes).ShouldNot(BeNil())
				Expect(status.Code(err)).Should(Equal(codes.OK))
			})
		})

		Describe("Deleting an longrunning with incorrect longrunning idand user id", func() {
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
