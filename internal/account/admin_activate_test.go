package account

import (
	"context"

	"github.com/gidyon/services/pkg/api/account"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Admin activating an account @admin_activate", func() {
	var (
		adminActivateReq *account.AdminUpdateAccountRequest
		ctx              context.Context
		adminActive      string
		adminInActive    string
		accountID        string
	)

	BeforeEach(func() {
		adminActivateReq = &account.AdminUpdateAccountRequest{
			AccountId:       uuid.New().String(),
			AdminId:         uuid.New().String(),
			UpdateOperation: account.UpdateOperation_ADMIN_ACTIVATE,
		}
		ctx = context.Background()
	})

	Describe("Calling AdminUpdateAccount admin activate with nil or malformed request", func() {
		It("should fail when the request is nil", func() {
			adminActivateReq = nil
			adminActivateRes, err := AccountAPI.AdminUpdateAccount(ctx, adminActivateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(adminActivateRes).Should(BeNil())
		})
		It("should fail when the admin id is missing", func() {
			adminActivateReq.AdminId = ""
			adminActivateRes, err := AccountAPI.AdminUpdateAccount(ctx, adminActivateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(adminActivateRes).Should(BeNil())
		})
		It("should fail when the account id is missing", func() {
			adminActivateReq.AccountId = ""
			adminActivateRes, err := AccountAPI.AdminUpdateAccount(ctx, adminActivateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(adminActivateRes).Should(BeNil())
		})
		It("should fail when operation is unknown", func() {
			adminActivateReq.UpdateOperation = account.UpdateOperation_UPDATE_OPERATION_INSPECIFIED
			adminActivateRes, err := AccountAPI.AdminUpdateAccount(ctx, adminActivateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(adminActivateRes).Should(BeNil())
		})
	})

	Describe("Creating Admins", func() {
		var err error
		It("should create admin", func() {
			adminActive, err = createAdmin(account.AccountState_ACTIVE)
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should create admin", func() {
			adminInActive, err = createAdmin(account.AccountState_INACTIVE)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("Calling AdminUpdateAccount with incorrect admin id or account id", func() {
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

			Context("When the admin id or account id is incorrect", func() {
				It("should fail when the admin id is incorrect", func() {
					adminActivateReq.AccountId = accountID
					adminActivateRes, err := AccountAPI.AdminUpdateAccount(ctx, adminActivateReq)
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
					Expect(adminActivateRes).Should(BeNil())
				})
				It("should fail when the account id is incorrect", func() {
					adminActivateReq.AdminId = adminActive
					adminActivateRes, err := AccountAPI.AdminUpdateAccount(ctx, adminActivateReq)
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
					Expect(adminActivateRes).Should(BeNil())
				})
				It("should succeed when the account id and admin id is correct", func() {
					adminActivateReq.AdminId = adminActive
					adminActivateReq.AccountId = accountID
					adminActivateRes, err := AccountAPI.AdminUpdateAccount(ctx, adminActivateReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(adminActivateRes).ShouldNot(BeNil())
				})
			})
		})
	})

	Describe("Calling AdminUpdateAccount account on existing account", func() {
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

			Describe("Admin activating the account", func() {
				Context("When the the admin account state is not ACTIVE", func() {
					It("should fail because the admin state is INACTIVE", func() {
						adminActivateReq.AccountId = accountID
						adminActivateReq.AdminId = adminInActive
						adminActivateRes, err := AccountAPI.AdminUpdateAccount(ctx, adminActivateReq)
						Expect(err).Should(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.PermissionDenied))
						Expect(adminActivateRes).Should(BeNil())
					})

					Describe("Let's get the account", func() {
						It("should succeed but account state still INACTIVE", func() {
							getRes, err := AccountAPI.GetAccount(ctx, &account.GetAccountRequest{
								AccountId: accountID,
							})
							Expect(err).ShouldNot(HaveOccurred())
							Expect(status.Code(err)).Should(Equal(codes.OK))
							Expect(getRes).ShouldNot(BeNil())
							Expect(getRes.State).Should(Equal(account.AccountState_INACTIVE))
						})
					})
				})

				Context("When the the account state ACTIVE", func() {
					It("should succeed because the admin state is ACTIVE", func() {
						adminActivateReq.AccountId = accountID
						adminActivateReq.AdminId = adminActive
						adminActivateRes, err := AccountAPI.AdminUpdateAccount(ctx, adminActivateReq)
						Expect(err).ShouldNot(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.OK))
						Expect(adminActivateRes).ShouldNot(BeNil())
					})

					Describe("Let's get the account", func() {
						It("should succeed and account state should be ACTIVE", func() {
							getRes, err := AccountAPI.GetAccount(ctx, &account.GetAccountRequest{
								AccountId: accountID,
							})
							Expect(err).ShouldNot(HaveOccurred())
							Expect(status.Code(err)).Should(Equal(codes.OK))
							Expect(getRes).ShouldNot(BeNil())
							Expect(getRes.State).Should(Equal(account.AccountState_ACTIVE))
						})
					})
				})
			})
		})
	})
})
