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

var _ = Describe("Deleting Account @delete", func() {
	var (
		delReq *account.DeleteAccountRequest
		ctx    context.Context
	)

	BeforeEach(func() {
		delReq = &account.DeleteAccountRequest{
			AccountId: uuid.New().String(),
		}
		ctx = context.Background()
	})

	Describe("Deleting account with malfored request", func() {
		It("should fail when request is nil", func() {
			delReq = nil
			delRes, err := AccountAPI.DeleteAccount(ctx, delReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(delRes).Should(BeNil())
		})
		It("should fail when account id is missing", func() {
			delReq.AccountId = ""
			delRes, err := AccountAPI.DeleteAccount(ctx, delReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(delRes).Should(BeNil())
		})
		It("should fail when account id incorrect", func() {
			delReq.AccountId = randomdata.RandStringRunes(20)
			delRes, err := AccountAPI.DeleteAccount(ctx, delReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(delRes).Should(BeNil())
		})
		It("should fail when account id doesn't exist", func() {
			delReq.AccountId = fmt.Sprint(randomdata.Number(2000000000, 10500000000))
			delRes, err := AccountAPI.DeleteAccount(ctx, delReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.NotFound))
			Expect(delRes).Should(BeNil())
		})
	})

	When("Deleting account with correct account id", func() {
		var accountID string
		Describe("Creating the account first", func() {
			It("should create account in database without error", func() {
				createRes, err := AccountAPI.CreateAccount(ctx, &account.CreateAccountRequest{
					Account:        fakeAccount(),
					PrivateAccount: fakePrivateAccount(),
					ProjectId:      "1",
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(createRes).ShouldNot(BeNil())
				accountID = createRes.AccountId
			})
			Describe("Deleting the account", func() {
				It("should delete account in database without error", func() {
					delReq.AccountId = accountID
					delRes, err := AccountAPI.DeleteAccount(ctx, delReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(delRes).ShouldNot(BeNil())
				})

				Describe("Getting the account", func() {
					It("should fail because it has been deleted", func() {
						getReq := &account.GetAccountRequest{
							AccountId: accountID,
						}
						getRes, err := AccountAPI.GetAccount(ctx, getReq)
						Expect(err).Should(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.NotFound))
						Expect(getRes).Should(BeNil())
					})
				})
			})
		})
	})
})
