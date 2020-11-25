package account

import (
	"context"

	"github.com/gidyon/services/pkg/api/account"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Blocking account £block", func() {
	var (
		blockReq      *account.AdminUpdateAccountRequest
		ctx           context.Context
		adminActive   string
		adminInActive string
		accountID     string
	)

	BeforeEach(func() {
		blockReq = &account.AdminUpdateAccountRequest{
			AccountId:       uuid.New().String(),
			AdminId:         uuid.New().String(),
			UpdateOperation: account.UpdateOperation_BLOCK,
		}
		ctx = context.Background()
	})

	Describe("Calling AdminUpdateAccount block with nil or malformed request", func() {
		It("should fail when the request is nil", func() {
			blockReq = nil
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, blockReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(blockRes).Should(BeNil())
		})
		It("should fail when the admin id is missing", func() {
			blockReq.AdminId = ""
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, blockReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(blockRes).Should(BeNil())
		})
		It("should fail when the account id is missing", func() {
			blockReq.AccountId = ""
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, blockReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(blockRes).Should(BeNil())
		})
		It("should fail when operation is unknown", func() {
			blockReq.UpdateOperation = account.UpdateOperation_UPDATE_OPERATION_INSPECIFIED
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, blockReq)
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

	Describe("Calling AdminUpdateAccount block with incorrect admin id or account id", func() {
		Context("Lets create an account first", func() {
			var accountID string
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

			Context("When the admin id or account id is incorrect", func() {
				It("should fail when the admin id is incorrect", func() {
					blockReq.AccountId = accountID
					blockRes, err := AccountAPI.AdminUpdateAccount(ctx, blockReq)
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
					Expect(blockRes).Should(BeNil())
				})
				It("should fail when the account id is incorrect", func() {
					blockReq.AdminId = adminActive
					blockRes, err := AccountAPI.AdminUpdateAccount(ctx, blockReq)
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
					Expect(blockRes).Should(BeNil())
				})
				It("should fail even when the account id and admin id is correct because account is not active", func() {
					blockReq.AdminId = adminActive
					blockReq.AccountId = accountID
					blockRes, err := AccountAPI.AdminUpdateAccount(ctx, blockReq)
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.FailedPrecondition))
					Expect(blockRes).Should(BeNil())
				})
			})
		})
	})

	Describe("Calling block account on existing account", func() {
		Describe("Creating user account", func() {
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
		})

		Context("When the the admin account state is not ACTIVE", func() {
			It("should fail because the admin state is INACTIVE", func() {
				blockReq.AccountId = accountID
				blockReq.AdminId = adminInActive
				blockRes, err := AccountAPI.AdminUpdateAccount(ctx, blockReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.PermissionDenied))
				Expect(blockRes).Should(BeNil())
			})

			Describe("Let's get the account", func() {
				It("should succeed because the account is not blocked", func() {
					getRes, err := AccountAPI.GetAccount(ctx, &account.GetAccountRequest{
						AccountId: accountID,
					})
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(getRes).ShouldNot(BeNil())
				})
			})
		})

		Context("When the the admin account state ACTIVE", func() {
			Describe("Lets activate the account in order to block it", func() {
				It("should succeed because the admin state is ACTIVE", func() {
					blockReq := &account.AdminUpdateAccountRequest{
						AccountId:       accountID,
						AdminId:         adminActive,
						UpdateOperation: account.UpdateOperation_ADMIN_ACTIVATE,
					}
					blockRes, err := AccountAPI.AdminUpdateAccount(ctx, blockReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(blockRes).ShouldNot(BeNil())
				})
			})

			It("should succeed because the admin state is ACTIVE", func() {
				blockReq.AccountId = accountID
				blockReq.AdminId = adminActive
				blockRes, err := AccountAPI.AdminUpdateAccount(ctx, blockReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(blockRes).ShouldNot(BeNil())
			})

			Describe("Let's get the account", func() {
				It("should fail because the account is blocked", func() {
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
})
