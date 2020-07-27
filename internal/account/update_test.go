package account

import (
	"context"
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/messaging"
	"github.com/gidyon/services/pkg/auth"

	"github.com/gidyon/services/pkg/api/account"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Update Account @update", func() {
	Describe("Updating public data @updatepublic", func() {
		var (
			updateReq *account.UpdateAccountRequest
			ctx       context.Context
		)

		BeforeEach(func() {
			updateReq = &account.UpdateAccountRequest{
				Account: fakeAccount(),
			}
			// Reset some fields
			updateReq.Account.Nationality = ""
			updateReq.Account.BirthDate = ""

			ctx = context.Background()
		})

		Describe("Updating account with malformed request", func() {
			It("should fail when request is nil", func() {
				updateReq = nil
				updateRes, err := AccountAPI.UpdateAccount(ctx, updateReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(updateRes).Should(BeNil())
			})
			It("should fail when account is nil", func() {
				updateReq.Account = nil
				updateRes, err := AccountAPI.UpdateAccount(ctx, updateReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(updateRes).Should(BeNil())
			})
			It("should definitely fail when account id is missing", func() {
				updateReq.Account.AccountId = ""
				updateRes, err := AccountAPI.UpdateAccount(ctx, updateReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(updateRes).Should(BeNil())
			})
			It("should definitely fail when account id is incorrect", func() {
				updateReq.Account.AccountId = "omen"
				updateRes, err := AccountAPI.UpdateAccount(ctx, updateReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(updateRes).Should(BeNil())
			})
			It("should definitely fail when account id is non-existence", func() {
				updateReq.Account.AccountId = fmt.Sprint(randomdata.Number(10000000, 100999999))
				updateRes, err := AccountAPI.UpdateAccount(ctx, updateReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.NotFound))
				Expect(updateRes).Should(BeNil())
			})
		})

		Context("Updating account with valid request", func() {
			var (
				accountGroup string
				accountID    string
			)

			Describe("Create account first", func() {
				It("should create account without error", func() {
					createReq := &account.CreateAccountRequest{
						Account:        fakeAccount(),
						PrivateAccount: fakePrivateAccount(),
					}
					// Create user account
					createReq.Account.Group = auth.User()
					createRes, err := AccountAPI.CreateAccount(ctx, createReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(createRes).ShouldNot(BeNil())
					accountID = createRes.AccountId
					accountGroup = createReq.Account.Group
				})
			})

			It("should update account in database without error", func() {
				updateReq.Account.AccountId = accountID
				// Set the account state to active
				updateReq.Account.State = account.AccountState_ACTIVE
				updateReq.Account.Group = auth.AdminGroup()
				updateRes, err := AccountAPI.UpdateAccount(ctx, updateReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(updateRes).ShouldNot(BeNil())
			})

			Describe("Getting the updated account", func() {
				It("should not be possible to have updated account state or account type", func() {
					getReq := &account.GetAccountRequest{
						AccountId: accountID,
					}
					getReq.AccountId = accountID
					getRes, err := AccountAPI.GetAccount(ctx, getReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(getRes).ShouldNot(BeNil())

					// The account state should remain unchanged
					Expect(getRes.State).Should(Equal(account.AccountState_INACTIVE))
					// The account type should remain unchanged
					Expect(getRes.Group).Should(Equal(accountGroup))
				})
			})
		})
	})
})

var _ = Describe("Updating private account @updateprivate", func() {
	var (
		updateReq *account.UpdatePrivateAccountRequest
		ctx       context.Context
	)

	BeforeEach(func() {
		updateReq = &account.UpdatePrivateAccountRequest{
			AccountId:      uuid.New().String(),
			PrivateAccount: fakePrivateAccount(),
			ChangeToken:    fmt.Sprint(randomdata.Number(100000, 999999)),
		}
		ctx = context.Background()
	})

	Describe("Updating account private profile with nil request", func() {
		It("should fail when request is nil", func() {
			updateReq = nil
			updateRes, err := AccountAPI.UpdatePrivateAccount(ctx, updateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(updateRes).Should(BeNil())
		})
		It("should definitely fail when change token is missing", func() {
			updateReq.ChangeToken = ""
			updateRes, err := AccountAPI.UpdatePrivateAccount(ctx, updateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(updateRes).Should(BeNil())
		})
		It("should definitely fail when account id is missing", func() {
			updateReq.AccountId = ""
			updateRes, err := AccountAPI.UpdatePrivateAccount(ctx, updateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(updateRes).Should(BeNil())
		})
		It("should definitely fail when account id is incorrect", func() {
			updateReq.AccountId = "omen"
			updateRes, err := AccountAPI.UpdatePrivateAccount(ctx, updateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(updateRes).Should(BeNil())
		})
		It("should definitely fail when account id is non-existening", func() {
			updateReq.AccountId = fmt.Sprint(randomdata.Number(1000000, 1000000000))
			updateRes, err := AccountAPI.UpdatePrivateAccount(ctx, updateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.NotFound))
			Expect(updateRes).Should(BeNil())
		})
		It("should fail when account is nil", func() {
			updateReq.PrivateAccount = nil
			updateRes, err := AccountAPI.UpdatePrivateAccount(ctx, updateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(updateRes).Should(BeNil())
		})
	})

	Describe("Create account first", func() {
		var (
			accountPB *account.Account
			accountID string
			token     string
		)
		It("should create account without error", func() {
			accountPB = fakeAccount()
			createRes, err := AccountAPI.CreateAccount(ctx, &account.CreateAccountRequest{
				Account:        accountPB,
				PrivateAccount: fakePrivateAccount(),
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			Expect(createRes).ShouldNot(BeNil())
			accountID = createRes.AccountId
		})

		BeforeEach(func() {
			updateReq.AccountId = accountID
		})

		Context("Asking for update token", func() {
			It("should request for token", func() {
				reqReq := &account.RequestChangePrivateAccountRequest{
					Payload:     accountPB.Email,
					FallbackUrl: randomdata.MacAddress(),
					SendMethod:  messaging.SendMethod_EMAIL,
				}
				updateRes, err := AccountAPI.RequestChangePrivateAccount(ctx, reqReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(updateRes).ShouldNot(BeNil())

				v, err := AccountAPIServer.redisDB.Get(updateToken(accountID)).Result()
				Expect(err).ShouldNot(HaveOccurred())
				token = v
			})

			Context("Updating private account with incorrect token", func() {
				It("should fail when token in icorrect", func() {
					updateRes, err := AccountAPI.UpdatePrivateAccount(ctx, updateReq)
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
					Expect(updateRes).Should(BeNil())
				})
			})
		})

		Describe("Updating account with update token", func() {
			BeforeEach(func() {
				updateReq.ChangeToken = token
			})

			Context("Updating account private profile with bad private payload", func() {
				It("should fail when private profile is nil", func() {
					updateReq.PrivateAccount = nil
					updateRes, err := AccountAPI.UpdatePrivateAccount(ctx, updateReq)
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
					Expect(updateRes).Should(BeNil())
				})
				It("should fail when passwords do not match", func() {
					updateReq.PrivateAccount.Password = "we dont match"
					updateRes, err := AccountAPI.UpdatePrivateAccount(ctx, updateReq)
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
					Expect(updateRes).Should(BeNil())
				})
			})

			Context("Updating account private profile with valid request", func() {
				It("should update account in database without error", func() {
					updateReq.PrivateAccount.Password = "hakty11"
					updateReq.PrivateAccount.ConfirmPassword = "hakty11"
					updateRes, err := AccountAPI.UpdatePrivateAccount(ctx, updateReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(updateRes).ShouldNot(BeNil())
				})
			})
		})
	})
})