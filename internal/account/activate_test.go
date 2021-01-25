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

const projectID = "test"

var _ = Describe("Activating user Account @activate", func() {
	var (
		activateReq *account.ActivateAccountRequest
		activateRes *account.ActivateAccountResponse
		ctx         context.Context
		err         error
		token       string
		accountID   string
		email       string
		password    string
	)

	BeforeEach(func() {
		activateReq = &account.ActivateAccountRequest{
			AccountId: uuid.New().String(),
			Token:     randomdata.RandStringRunes(32),
		}
		ctx = context.Background()
	})

	When("Activating account with missing or nil request", func() {
		It("should fail when request is nil", func() {
			activateReq = nil
			activateRes, err = AccountAPI.ActivateAccount(ctx, activateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(activateRes).Should(BeNil())
		})
		It("should fail when token is missing", func() {
			activateReq.Token = ""
			activateRes, err = AccountAPI.ActivateAccount(ctx, activateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(activateRes).Should(BeNil())
		})
		It("should fail when account id is missing", func() {
			activateReq.AccountId = ""
			activateRes, err = AccountAPI.ActivateAccount(ctx, activateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(activateRes).Should(BeNil())
		})
		It("should fail when account id is incorrect", func() {
			activateRes, err = AccountAPI.ActivateAccount(ctx, activateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(activateRes).Should(BeNil())
		})
		It("should fail when account id is non-existence", func() {
			activateReq.AccountId = fmt.Sprint(randomdata.Number(100000000, 1099999999))
			activateRes, err = AccountAPI.ActivateAccount(ctx, activateReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.NotFound))
			Expect(activateRes).Should(BeNil())
		})
	})

	Context("Activating account => create account, signIn to get token and id, then activate", func() {
		var group string
		Describe("Creating an account", func() {

			It("should create account in database without error", func() {
				createReq := &account.CreateAccountRequest{
					Account:        fakeAccount(),
					PrivateAccount: fakePrivateAccount(),
					ProjectId:      projectID,
				}

				createRes, err := AccountAPI.CreateAccount(ctx, createReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(createRes).ShouldNot(BeNil())
				Expect(createRes.AccountId).ShouldNot(BeZero())

				email = createReq.Account.Email
				password = createReq.PrivateAccount.Password
				accountID = createRes.AccountId
				group = createReq.Account.Group
			})

			Describe("SignIn to the created account", func() {
				It("should signIn the account and return some data", func() {
					signInReq := &account.SignInRequest{
						Username:  email,
						Password:  password,
						Group:     group,
						ProjectId: projectID,
					}
					signInRes, err := AccountAPI.SignIn(ctx, signInReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(signInRes).ShouldNot(BeNil())
					Expect(signInRes.Token).ShouldNot(BeZero())
					Expect(signInRes.AccountId).ShouldNot(BeZero())

					token = signInRes.Token
				})

				Describe("Activating the account", func() {
					It("should activate the account in database", func() {
						activateReq.AccountId = accountID
						activateReq.Token = token
						activateRes, err = AccountAPI.ActivateAccount(ctx, activateReq)
						Expect(err).ShouldNot(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.OK))
						Expect(activateRes).ShouldNot(BeNil())
					})

					Describe("Getting account the account and checking if its activated", func() {
						It("should get the account", func() {
							activeState := account.AccountState_ACTIVE.String()
							getReq := &account.GetAccountRequest{
								AccountId: accountID,
							}
							getRes, err := AccountAPI.GetAccount(ctx, getReq)
							Expect(err).ShouldNot(HaveOccurred())
							Expect(status.Code(err)).Should(Equal(codes.OK))
							Expect(getRes).ShouldNot(BeNil())
							Expect(getRes.State.String()).Should(BeEquivalentTo(activeState))
						})
					})
				})
			})
		})
	})
})
