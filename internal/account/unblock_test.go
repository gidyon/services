package account

import (
	"context"

	"github.com/gidyon/services/pkg/api/account"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Unblocking an account @unblock", func() {
	var (
		unBlockReq    *account.AdminUpdateAccountRequest
		ctx           context.Context
		adminActive   string
		adminInActive string
		accountID     string
	)

	BeforeEach(func() {
		unBlockReq = &account.AdminUpdateAccountRequest{
			AccountId:       uuid.New().String(),
			AdminId:         uuid.New().String(),
			UpdateOperation: account.UpdateOperation_UNBLOCK,
		}
		ctx = context.Background()
	})

	Describe("Calling AdminUpdateAccount with nil or malformed request", func() {
		It("should fail when the request is nil", func() {
			unBlockReq = nil
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, unBlockReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(blockRes).Should(BeNil())
		})
		It("should fail when the admin id is missing", func() {
			unBlockReq.AdminId = ""
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, unBlockReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(blockRes).Should(BeNil())
		})
		It("should fail when the account id is missing", func() {
			unBlockReq.AccountId = ""
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, unBlockReq)
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

	Describe("Calling unblock account on existing account", func() {
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

			Describe("Unblocking an account when the account is not blocked", func() {
				It("should fail to unblock the account because it is not blocked", func() {
					unBlockReq.AccountId = accountID
					unBlockReq.AdminId = adminActive
					blockRes, err := AccountAPI.AdminUpdateAccount(ctx, unBlockReq)
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.FailedPrecondition))
					Expect(blockRes).Should(BeNil())
				})

			})

			Describe("Blocking an account", func() {
				Context("Lets activate the account then block it, then unblock it", func() {
					Describe("Lets activate the account in order to block it", func() {
						It("should succeed because the admin state is ACTIVE", func() {
							blockReq := &account.AdminUpdateAccountRequest{
								AccountId:       accountID,
								AdminId:         adminActive,
								UpdateOperation: account.UpdateOperation_ADMIN_ACTIVATE,
							}
							adminActivateRes, err := AccountAPI.AdminUpdateAccount(ctx, blockReq)
							Expect(err).ShouldNot(HaveOccurred())
							Expect(status.Code(err)).Should(Equal(codes.OK))
							Expect(adminActivateRes).ShouldNot(BeNil())
						})
					})

					Describe("Lets block the account now", func() {
						It("should succeed in blocking the account because the admin state is ACTIVE", func() {
							blockReq := &account.AdminUpdateAccountRequest{
								AccountId:       accountID,
								AdminId:         adminActive,
								UpdateOperation: account.UpdateOperation_BLOCK,
							}
							adminActivateRes, err := AccountAPI.AdminUpdateAccount(ctx, blockReq)
							Expect(err).ShouldNot(HaveOccurred())
							Expect(status.Code(err)).Should(Equal(codes.OK))
							Expect(adminActivateRes).ShouldNot(BeNil())
						})
					})

					Describe("Unblocking the account", func() {

						Context("When the the admin account state is not ACTIVE", func() {
							Describe("Lets try to unblock the account", func() {
								It("should fail because the admin state is INACTIVE", func() {
									unBlockReq.AccountId = accountID
									unBlockReq.AdminId = adminInActive
									blockRes, err := AccountAPI.AdminUpdateAccount(ctx, unBlockReq)
									Expect(err).Should(HaveOccurred())
									Expect(status.Code(err)).Should(Equal(codes.PermissionDenied))
									Expect(blockRes).Should(BeNil())
								})

								Describe("Let's get the account", func() {
									It("should fail because the account is still blocked", func() {
										getRes, err := AccountAPI.GetAccount(ctx, &account.GetAccountRequest{
											AccountId: accountID,
										})
										Expect(err).Should(HaveOccurred())
										Expect(status.Code(err)).Should(Equal(codes.PermissionDenied))
										Expect(getRes).Should(BeNil())
									})
								})
							})
						})

						Context("When the the admin account state ACTIVE", func() {
							Describe("Lets try to unblock the account", func() {
								It("should succeed because the admin state is ACTIVE", func() {
									unBlockReq.AccountId = accountID
									unBlockReq.AdminId = adminActive
									blockRes, err := AccountAPI.AdminUpdateAccount(ctx, unBlockReq)
									Expect(err).ShouldNot(HaveOccurred())
									Expect(status.Code(err)).Should(Equal(codes.OK))
									Expect(blockRes).ShouldNot(BeNil())
								})

								Describe("Let's get the account", func() {
									It("should succeed because the account has been unblocked", func() {
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
})
