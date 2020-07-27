package account

import (
	"context"

	"github.com/gidyon/services/pkg/api/account"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Change Account Type @change", func() {
	var (
		changeReq     *account.AdminUpdateAccountRequest
		ctx           context.Context
		adminActive   string
		adminInActive string
		accountID     string
	)

	BeforeEach(func() {
		changeReq = &account.AdminUpdateAccountRequest{
			AccountId:       uuid.New().String(),
			AdminId:         uuid.New().String(),
			UpdateOperation: account.UpdateOperation_CHANGE_GROUP,
		}
		ctx = context.Background()
	})

	Context("Failure scenarios", func() {
		When("Changing account type", func() {
			It("should fail when the request is nil", func() {
				changeReq = nil
				changeRes, err := AccountAPI.AdminUpdateAccount(ctx, changeReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(changeRes).Should(BeNil())
			})
			It("should fail when the account id is missing", func() {
				changeReq.AccountId = ""
				changeRes, err := AccountAPI.AdminUpdateAccount(ctx, changeReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(changeRes).Should(BeNil())
			})
			It("should fail when the admin id is missing", func() {
				changeReq.AdminId = ""
				changeRes, err := AccountAPI.AdminUpdateAccount(ctx, changeReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(changeRes).Should(BeNil())
			})
			It("should fail when operation is unknown", func() {
				changeReq.UpdateOperation = account.UpdateOperation_UPDATE_OPERATION_INSPECIFIED
				changeRes, err := AccountAPI.AdminUpdateAccount(ctx, changeReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(changeRes).Should(BeNil())
			})
		})
	})

	Describe("Creating Admins", func() {
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

	Describe("Calling AdminUpdateAccount with incorrect admin id or account id", func() {
		Context("Lets create an account first", func() {
			It("should create the account without any error", func() {
				createReq := &account.CreateAccountRequest{
					Account:        fakeAccount(),
					PrivateAccount: fakePrivateAccount(),
				}
				createRes, err := AccountAPI.CreateAccount(ctx, createReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(createRes).ShouldNot(BeNil())
				accountID = createRes.AccountId
			})

			Context("When the admin id or account id is incorrect", func() {
				It("should fail when the admin id is incorrect", func() {
					changeReq.AccountId = accountID
					adminActivateRes, err := AccountAPI.AdminUpdateAccount(ctx, changeReq)
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
					Expect(adminActivateRes).Should(BeNil())
				})
				It("should fail when the account id is incorrect", func() {
					changeReq.AdminId = adminActive
					adminActivateRes, err := AccountAPI.AdminUpdateAccount(ctx, changeReq)
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
					Expect(adminActivateRes).Should(BeNil())
				})
				It("should fail even when the account id and admin id is correct becuase account is still inactive", func() {
					changeReq.AdminId = adminActive
					changeReq.AccountId = accountID
					adminActivateRes, err := AccountAPI.AdminUpdateAccount(ctx, changeReq)
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.FailedPrecondition))
					Expect(adminActivateRes).Should(BeNil())
				})
			})
		})
	})

	Describe("Calling AdminUpdateAccount account on existing account", func() {

		Context("Lets create an account first", func() {
			var accountID string
			It("should create the account without any error", func() {
				createReq := &account.CreateAccountRequest{
					Account:        fakeAccount(),
					PrivateAccount: fakePrivateAccount(),
				}
				createRes, err := AccountAPI.CreateAccount(ctx, createReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(createRes).ShouldNot(BeNil())
				accountID = createRes.AccountId
			})

			Describe("Admin changing the account type", func() {

				Context("When the the admin account state is not ACTIVE", func() {
					It("should fail because the admin state is INACTIVE", func() {
						changeReq.AccountId = accountID
						changeReq.AdminId = adminInActive
						changeRes, err := AccountAPI.AdminUpdateAccount(ctx, changeReq)
						Expect(err).Should(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.PermissionDenied))
						Expect(changeRes).Should(BeNil())
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
						changeReq.AccountId = accountID
						changeReq.AdminId = adminActive
						changeRes, err := AccountAPI.AdminUpdateAccount(ctx, changeReq)
						Expect(err).ShouldNot(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.OK))
						Expect(changeRes).ShouldNot(BeNil())
					})

					Context("Let's get the account", func() {
						It("should succeed and account type changed", func() {
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
