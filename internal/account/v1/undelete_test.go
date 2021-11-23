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

var _ = Describe("Restoring a deleted account @undelete", func() {
	var (
		undeleteReq   *account.AdminUpdateAccountRequest
		ctx           context.Context
		adminActive   string
		adminInActive string
		accountID     string
	)

	BeforeEach(func() {
		undeleteReq = &account.AdminUpdateAccountRequest{
			AccountId:       uuid.New().String(),
			AdminId:         uuid.New().String(),
			UpdateOperation: account.UpdateOperation_UNDELETE,
		}
		ctx = context.Background()
	})

	Describe("Calling AdminUpdateAccount with nil or malformed request", func() {
		It("should fail when the request is nil", func() {
			undeleteReq = nil
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, undeleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(blockRes).Should(BeNil())
		})
		It("should fail when the admin id is missing", func() {
			undeleteReq.AdminId = ""
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, undeleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(blockRes).Should(BeNil())
		})
		It("should fail when the account id is missing", func() {
			undeleteReq.AccountId = ""
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, undeleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(blockRes).Should(BeNil())
		})
		It("should fail when the operation is unknown", func() {
			undeleteReq.UpdateOperation = account.UpdateOperation_UPDATE_OPERATION_INSPECIFIED
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, undeleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(blockRes).Should(BeNil())
		})
	})

	Describe("Creating admins", func() {
		var err error
		It("should create active admin", func() {
			adminActive, err = createAdmin(account.AccountState_ACTIVE)
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should create inactive admin", func() {
			adminInActive, err = createAdmin(account.AccountState_INACTIVE)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("Calling undelete account on existing account", func() {

		Context("Lets create an account first", func() {
			It("should create the account without any error", func() {
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

			Describe("Deleting an account", func() {
				It("should delete the account", func() {
					undeleteReq.AccountId = accountID
					undeleteReq.AdminId = adminActive
					blockRes, err := AccountAPI.DeleteAccount(ctx, &account.DeleteAccountRequest{
						AccountId: accountID,
					})
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(blockRes).ShouldNot(BeNil())
				})

				Describe("Restoring deleted the account", func() {

					Context("When the the admin and account id is incorrect", func() {
						It("should fail when account id is incorrect", func() {
							undeleteReq.AccountId = fmt.Sprint(randomdata.Number(1234567, 76543218))
							undeleteReq.AdminId = adminActive
							blockRes, err := AccountAPI.AdminUpdateAccount(ctx, undeleteReq)
							Expect(err).Should(HaveOccurred())
							Expect(status.Code(err)).Should(Equal(codes.NotFound))
							Expect(blockRes).Should(BeNil())
						})

						It("should fail when admin id is incorrect", func() {
							undeleteReq.AccountId = accountID
							undeleteReq.AdminId = fmt.Sprint(randomdata.Number(1234567, 76543218))
							blockRes, err := AccountAPI.AdminUpdateAccount(ctx, undeleteReq)
							Expect(err).Should(HaveOccurred())
							Expect(status.Code(err)).Should(Equal(codes.NotFound))
							Expect(blockRes).Should(BeNil())
						})

						Describe("Let's get the account", func() {
							It("should fail because the account is still deleted", func() {
								getRes, err := AccountAPI.GetAccount(ctx, &account.GetAccountRequest{
									AccountId: accountID,
								})
								Expect(err).Should(HaveOccurred())
								Expect(status.Code(err)).Should(Equal(codes.NotFound))
								Expect(getRes).Should(BeNil())
							})
						})
					})

					Context("When the the admin account state is not ACTIVE", func() {
						It("should fail because the admin state is INACTIVE", func() {
							undeleteReq.AccountId = accountID
							undeleteReq.AdminId = adminInActive
							blockRes, err := AccountAPI.AdminUpdateAccount(ctx, undeleteReq)
							Expect(err).Should(HaveOccurred())
							Expect(status.Code(err)).Should(Equal(codes.PermissionDenied))
							Expect(blockRes).Should(BeNil())
						})

						Describe("Let's get the account", func() {
							It("should fail because the account is still deleted", func() {
								getRes, err := AccountAPI.GetAccount(ctx, &account.GetAccountRequest{
									AccountId: accountID,
								})
								Expect(err).Should(HaveOccurred())
								Expect(status.Code(err)).Should(Equal(codes.NotFound))
								Expect(getRes).Should(BeNil())
							})
						})
					})

					Context("When the the admin account state ACTIVE", func() {
						Describe("Lets try to undelete the account", func() {
							It("should succeed because the admin state is ACTIVE", func() {
								undeleteReq.AccountId = accountID
								undeleteReq.AdminId = adminActive
								blockRes, err := AccountAPI.AdminUpdateAccount(ctx, undeleteReq)
								Expect(err).ShouldNot(HaveOccurred())
								Expect(status.Code(err)).Should(Equal(codes.OK))
								Expect(blockRes).ShouldNot(BeNil())
							})

							Describe("Let's get the account", func() {
								It("should succeed because the account has been undeleted", func() {
									getRes, err := AccountAPI.GetAccount(ctx, &account.GetAccountRequest{
										AccountId: accountID,
									})
									Expect(err).ShouldNot(HaveOccurred())
									Expect(status.Code(err)).Should(Equal(codes.OK))
									Expect(getRes).ShouldNot(BeNil())
								})
							})
						})
					})
				})
			})
		})
	})
})
