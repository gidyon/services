package longrunning

import (
	"context"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/longrunning"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Getting A Operation @get", func() {
	var (
		userID = randomdata.RandStringRunes(32)
		getReq *longrunning.GetOperationRequest
		ctx    context.Context
	)

	BeforeEach(func() {
		getReq = &longrunning.GetOperationRequest{
			OperationId: randomdata.RandStringRunes(32),
		}
		ctx = context.Background()
	})

	Describe("Getting an longrunning with incorrect/missing values", func() {
		It("should fail when the request is nil", func() {
			getReq = nil
			getRes, err := OperationAPIService.GetOperation(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(getRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when longrunning id is missing", func() {
			getReq.OperationId = ""
			getRes, err := OperationAPIService.GetOperation(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(getRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when longrunning id is incorrect", func() {
			getRes, err := OperationAPIService.GetOperation(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(getRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.NotFound))
		})
	})

	Describe("Getting a longrunning with correct/valid request", func() {
		var longrunningID string
		Describe("Lets create the longrunning first", func() {
			Describe("Creating an longrunning", func() {
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
			})

			Describe("Getting the longrunning", func() {
				It("should succeed when the request is valid", func() {
					getReq.OperationId = longrunningID
					getRes, err := OperationAPIService.GetOperation(ctx, getReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(getRes).ShouldNot(BeNil())
					Expect(status.Code(err)).Should(Equal(codes.OK))
				})
			})

			Describe("Checking whether the longrunning exists in cache", func() {
				It("should exist in cache", func() {
					opStr, err := OperationAPIService.RedisClient.Get(ctx, getOpKey(longrunningID)).Result()
					Expect(opStr).ShouldNot(BeZero())
					Expect(err).ShouldNot(HaveOccurred())
				})
			})
		})
	})
})
