package account

import (
	"context"
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/account"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Get Account @get", func() {
	var (
		getReq *account.GetAccountRequest
		ctx    context.Context
	)

	BeforeEach(func() {
		getReq = &account.GetAccountRequest{
			AccountId: uuid.New().String(),
		}
		ctx = context.Background()
	})

	Context("Get account with nil request", func() {
		It("should fail when request is nil", func() {
			getReq = nil
			getRes, err := AccountAPI.GetAccount(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(getRes).Should(BeNil())
		})
	})

	Context("Get account with missing/incorrect account id", func() {
		It("should fail when account id is missing", func() {
			getReq.AccountId = ""
			getRes, err := AccountAPI.GetAccount(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(getRes).Should(BeNil())
		})
		It("should fail when account id is incorrect", func() {
			getReq.AccountId = "knowledge"
			getRes, err := AccountAPI.GetAccount(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(getRes).Should(BeNil())
		})
		It("should fail when account id is non-exitence", func() {
			getReq.AccountId = fmt.Sprint(randomdata.Number(882111111, 999998888))
			getRes, err := AccountAPI.GetAccount(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.NotFound))
			Expect(getRes).Should(BeNil())
		})
	})

	Describe("Creating an account and getting it", func() {
		var accountID string
		It("should succeed in creating account", func() {
			createReq := &account.CreateAccountRequest{
				Account:        fakeAccount(),
				PrivateAccount: fakePrivateAccount(),
				ProjectId:      "1",
			}
			createRes, err := AccountAPI.CreateAccount(ctx, createReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			Expect(createRes).ShouldNot(BeNil())
			accountID = createRes.AccountId
		})

		Context("Get account with valid request", func() {
			It("should get the account", func() {
				getReq.AccountId = accountID
				getRes, err := AccountAPI.GetAccount(ctx, getReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(getRes).ShouldNot(BeNil())
			})
		})
	})
})
