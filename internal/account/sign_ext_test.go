package account

import (
	"context"
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/account"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Sign in a user @signin", func() {
	var (
		signInReq *account.SignInExternalRequest
		ctx       context.Context
	)

	BeforeEach(func() {
		signInReq = &account.SignInExternalRequest{
			Account: fakeAccount(),
			AuthToken: fmt.Sprintf(
				"%s.%s.%s",
				randomdata.RandStringRunes(32), randomdata.RandStringRunes(32), randomdata.RandStringRunes(32),
			),
			ProjectId: projectID,
		}
		ctx = context.Background()
	})

	Describe("Signing a user with malformed request", func() {
		It("should fail when the request is nil", func() {
			signInReq = nil
			signInRes, err := AccountAPI.SignInExternal(ctx, signInReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(signInRes).Should(BeNil())
		})
		It("should fail when the account information is missing", func() {
			signInReq.Account = nil
			signInRes, err := AccountAPI.SignInExternal(ctx, signInReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(signInRes).Should(BeNil())
		})
		It("should fail when names is missing", func() {
			signInReq.Account.Names = ""
			signInRes, err := AccountAPI.SignInExternal(ctx, signInReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(signInRes).Should(BeNil())
		})
		It("should fail when email and phone is missing", func() {
			signInReq.Account.Email = ""
			signInReq.Account.Phone = ""
			signInRes, err := AccountAPI.SignInExternal(ctx, signInReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(signInRes).Should(BeNil())
		})
		It("should fail when auth token is missing", func() {
			signInReq.AuthToken = ""
			signInRes, err := AccountAPI.SignInExternal(ctx, signInReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(signInRes).Should(BeNil())
		})
	})

	Describe("Signing a user with valid request", func() {
		Describe("Signing user", func() {
			It("should succeed when email is missing but phone is not", func() {
				signInReq.Account.Email = ""
				signInRes, err := AccountAPI.SignInExternal(ctx, signInReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(signInRes).ShouldNot(BeNil())
			})
			It("should succeed when phone is missing but email is not", func() {
				signInReq.Account.Phone = ""
				signInRes, err := AccountAPI.SignInExternal(ctx, signInReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(signInRes).ShouldNot(BeNil())
			})
		})

		var userID, names string
		It("should succeed when the request is valid", func() {
			signInRes, err := AccountAPI.SignInExternal(ctx, signInReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			Expect(signInRes).ShouldNot(BeNil())
			userID = signInRes.AccountId
			names = signInReq.Account.Names
		})

		Describe("Getting the new user", func() {
			var userPB *account.Account
			It("should return the user", func() {
				getRes, err := AccountAPI.GetAccount(ctx, &account.GetAccountRequest{
					AccountId: userID,
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(getRes).ShouldNot(BeNil())
				Expect(names).Should(Equal(getRes.Names))
				userPB = getRes
			})

			Describe("Signing user who already exists", func() {
				It("should succeed", func() {
					signInReq.Account = userPB
					signInRes, err := AccountAPI.SignInExternal(ctx, signInReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(signInRes).ShouldNot(BeNil())
				})

				It("should succeed if email is missing but phone is not", func() {
					signInReq.Account = userPB
					signInReq.Account.Email = ""
					signInRes, err := AccountAPI.SignInExternal(ctx, signInReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(signInRes).ShouldNot(BeNil())
				})
			})
		})
	})
})
