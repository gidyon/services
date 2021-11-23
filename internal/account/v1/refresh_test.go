package account

import (
	"context"
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/account"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Refreshing JWT @refresh", func() {
	var (
		refreshReq *account.RefreshSessionRequest
		ctx        context.Context
	)

	BeforeEach(func() {
		refreshReq = &account.RefreshSessionRequest{
			RefreshToken: randomdata.RandStringRunes(64),
			AccountGroup: randomdata.Adjective(),
			AccountId:    fmt.Sprint(randomdata.Number(1, 300)),
		}
		ctx = context.Background()
	})

	Describe("Calling RefreshSession with malformed request", func() {
		It("should fail when the request is nil", func() {
			refreshReq = nil
			refreshRes, err := AccountAPI.RefreshSession(ctx, refreshReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(refreshRes).Should(BeNil())
		})
		It("should fail when the refresh token is missing", func() {
			refreshReq.RefreshToken = ""
			refreshRes, err := AccountAPI.RefreshSession(ctx, refreshReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(refreshRes).Should(BeNil())
		})
		It("should fail when the account id is missing", func() {
			refreshReq.AccountId = ""
			refreshRes, err := AccountAPI.RefreshSession(ctx, refreshReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(refreshRes).Should(BeNil())
		})
		It("should fail when the account id is incorrect", func() {
			refreshReq.AccountId = randomdata.RandStringRunes(32)
			refreshRes, err := AccountAPI.RefreshSession(ctx, refreshReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(refreshRes).Should(BeNil())
		})
		It("should fail when the refresh token does not exist", func() {
			refreshReq.RefreshToken = randomdata.RandStringRunes(32)
			refreshRes, err := AccountAPI.RefreshSession(ctx, refreshReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.Unauthenticated))
			Expect(refreshRes).Should(BeNil())
		})
	})

	Describe("Calling RefreshSession with valid request", func() {
		Context("Let's create an account first", func() {
			var accountID, email, password, group string
			It("should create an account without error", func() {
				createReq := &account.CreateAccountRequest{
					Account:        fakeAccount(),
					PrivateAccount: fakePrivateAccount(),
					ProjectId:      projectID,
				}
				createRes, err := AccountAPI.CreateAccount(ctx, createReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(createRes).ShouldNot(BeNil())
				email = createReq.Account.Email
				password = createReq.PrivateAccount.Password
				accountID = createRes.AccountId
				group = createReq.Account.Group
			})
			Context("SignIng In into the account", func() {
				var signRes *account.SignInResponse
				It("should signIn into the account returning JWT and some data", func() {
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
					Expect(signInRes.AccountId).Should(Equal(accountID))
					signRes = signInRes
				})

				Describe("Refreshing jwt", func() {
					It("should succeed because the user is signed in", func() {
						refreshReq.AccountGroup = group
						refreshReq.AccountId = accountID
						refreshReq.RefreshToken = signRes.RefreshToken
						refreshRes, err := AccountAPI.RefreshSession(ctx, refreshReq)
						Expect(err).ShouldNot(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.OK))
						Expect(refreshRes).ShouldNot(BeNil())
					})
				})
			})
		})
	})
})
