package longrunning

import (
	"context"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/longrunning"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// List longrunning for incorrect user

var _ = Describe("ListOperations @list", func() {
	var (
		userID  = randomdata.RandStringRunes(32)
		listReq *longrunning.ListOperationsRequest
		ctx     context.Context
	)

	BeforeEach(func() {
		listReq = &longrunning.ListOperationsRequest{
			PageToken: "",
			PageSize:  10,
			Filter: &longrunning.ListOperationsFilter{
				UserId: userID,
			},
		}
		ctx = context.Background()
	})

	Describe("Calling ListOperations with missing/incorrect values", func() {
		It("should fail when the request is nil", func() {
			listReq = nil
			listRes, err := OperationAPI.ListOperations(ctx, listReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(listRes).Should(BeNil())
		})
		It("should fail when page token is icorrect request is nil", func() {
			listReq.PageToken = "nil"
			listRes, err := OperationAPI.ListOperations(ctx, listReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(listRes).Should(BeNil())
		})
		It("should fail when user id is missing", func() {
			listReq.Filter.UserId = ""
			listRes, err := OperationAPI.ListOperations(ctx, listReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(listRes).Should(BeNil())
		})
	})

	Describe("Calling ListOperations with valid values", func() {
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
				})
			}
		})

		It("should succeed when the request is valid", func() {
			listRes, err := OperationAPI.ListOperations(ctx, listReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			Expect(listRes).ShouldNot(BeNil())
			Expect(len(listRes.Operations)).ShouldNot(BeZero())
		})

		It("should succeed even when the page size is too big", func() {
			listReq.PageSize = 100
			listRes, err := OperationAPI.ListOperations(ctx, listReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			Expect(listRes).ShouldNot(BeNil())
		})
	})

	Describe("Listing longrunnings with incorrect user id", func() {
		It("should succeed but with no ops!", func() {
			listReq.Filter.UserId = randomdata.RandStringRunes(32)
			listRes, err := OperationAPI.ListOperations(ctx, listReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			Expect(listRes).ShouldNot(BeNil())
			Expect(len(listRes.Operations)).Should(BeZero())
		})
	})
})
