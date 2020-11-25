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

var _ = Describe("Admin deleting an account @admindelete", func() {
	var (
		deleteReq     *account.AdminUpdateAccountRequest
		ctx           context.Context
		adminActive   string
		adminInActive string
		accountID     string
	)

	BeforeEach(func() {
		deleteReq = &account.AdminUpdateAccountRequest{
			AccountId:       uuid.New().String(),
			AdminId:         uuid.New().String(),
			UpdateOperation: account.UpdateOperation_DELETE,
		}
		ctx = context.Background()
	})

	Describe("Calling AdminUpdateAccount with nil or malformed request", func() {
		It("should fail when the request is nil", func() {
			deleteReq = nil
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, deleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(blockRes).Should(BeNil())
		})
		It("should fail when the admin id is missing", func() {
			deleteReq.AdminId = ""
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, deleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(blockRes).Should(BeNil())
		})
		It("should fail when the account id is missing", func() {
			deleteReq.AccountId = ""
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, deleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(blockRes).Should(BeNil())
		})
		It("should fail when the operation is unknown", func() {
			deleteReq.UpdateOperation = account.UpdateOperation_UPDATE_OPERATION_INSPECIFIED
			blockRes, err := AccountAPI.AdminUpdateAccount(ctx, deleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(blockRes).Should(BeNil())
		})
	})

	Describe("Creating the admins", func() {
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

	Describe("Calling delete account on existing account", func() {
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

			Describe("Deleting the account", func() {
				Context("When the the admin and account id is incorrect", func() {
					It("should fail when account id is incorrect", func() {
						deleteReq.AccountId = fmt.Sprint(randomdata.Number(1234567, 76543218))
						deleteReq.AdminId = adminActive
						blockRes, err := AccountAPI.AdminUpdateAccount(ctx, deleteReq)
						Expect(err).Should(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.NotFound))
						Expect(blockRes).Should(BeNil())
					})

					It("should fail when admin id is incorrect", func() {
						deleteReq.AccountId = accountID
						deleteReq.AdminId = fmt.Sprint(randomdata.Number(1234567, 76543218))
						blockRes, err := AccountAPI.AdminUpdateAccount(ctx, deleteReq)
						Expect(err).Should(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.NotFound))
						Expect(blockRes).Should(BeNil())
					})

					Describe("Let's get the account", func() {
						It("should succeed because the account is still present", func() {
							getRes, err := AccountAPI.GetAccount(ctx, &account.GetAccountRequest{
								AccountId: accountID,
							})
							Expect(err).ShouldNot(HaveOccurred())
							Expect(status.Code(err)).Should(Equal(codes.OK))
							Expect(getRes).ShouldNot(BeNil())
						})
					})
				})

				Context("When the the admin account state is not ACTIVE", func() {
					It("should fail because the admin state is INACTIVE", func() {
						deleteReq.AccountId = accountID
						deleteReq.AdminId = adminInActive
						blockRes, err := AccountAPI.AdminUpdateAccount(ctx, deleteReq)
						Expect(err).Should(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.PermissionDenied))
						Expect(blockRes).Should(BeNil())
					})

					Describe("Let's get the account", func() {
						It("should succeed because the account is still present", func() {
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
					Describe("Lets try to delete the account", func() {
						It("should succeed because the admin state is ACTIVE", func() {
							deleteReq.AccountId = accountID
							deleteReq.AdminId = adminActive
							blockRes, err := AccountAPI.AdminUpdateAccount(ctx, deleteReq)
							Expect(err).ShouldNot(HaveOccurred())
							Expect(status.Code(err)).Should(Equal(codes.OK))
							Expect(blockRes).ShouldNot(BeNil())
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
				})
			})
		})
	})
})
