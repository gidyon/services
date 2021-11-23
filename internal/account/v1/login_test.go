package account

import (
	"context"
	"math/rand"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/services/pkg/api/account"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var groups = []string{auth.DefaultAdminGroup(), auth.DefaultSuperAdminGroup(), auth.DefaultUserGroup()}

func getGroup() string {
	return groups[rand.Intn(len(groups))]
}

var _ = Describe("SignIn Account @signIn", func() {
	var (
		signInReq *account.SignInRequest
		ctx       context.Context
	)

	BeforeEach(func() {
		signInReq = &account.SignInRequest{
			Username:  randomdata.Email(),
			Password:  randomdata.RandStringRunes(10),
			Group:     getGroup(),
			ProjectId: "test",
		}
		ctx = context.Background()
	})

	When("An account signIn with nil request", func() {
		It("should definitely fail", func() {
			signInReq = nil
			signInRes, err := AccountAPI.SignIn(ctx, signInReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(signInRes).Should(BeNil())
		})
	})

	When("A account signIn with missing credentials", func() {
		It("should fail when email and phone is missing in the signIn credentials", func() {
			signInReq.Username = ""
			signInRes, err := AccountAPI.SignIn(ctx, signInReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(signInRes).Should(BeNil())
		})
		It("should fail when password is missing in the signIn credentials", func() {
			signInReq.Password = ""
			signInRes, err := AccountAPI.SignIn(ctx, signInReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(signInRes).Should(BeNil())
		})
		It("should fail when project id is missing in the signIn credentials", func() {
			signInReq.ProjectId = ""
			signInRes, err := AccountAPI.SignIn(ctx, signInReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(signInRes).Should(BeNil())
		})
	})

	When("An account signIn with incorrect credentials", func() {
		It("should fail when email is incorrect", func() {
			signInReq.Username = "incorrect@gmail.com"
			signInRes, err := AccountAPI.SignIn(ctx, signInReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.NotFound))
			Expect(signInRes).Should(BeNil())
			Expect(err.Error()).Should(ContainSubstring("account"))
		})
		It("should fail when phone is incorrect", func() {
			signInReq.Username = "07notexist"
			signInRes, err := AccountAPI.SignIn(ctx, signInReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.NotFound))
			Expect(signInRes).Should(BeNil())
			Expect(err.Error()).Should(ContainSubstring("account"))
		})
	})

	When("A account signIn with valid credentials", func() {
		var accountID, email, password, group string
		Context("Let's create an account first", func() {
			It("should create account without error", func() {
				createReq := &account.CreateAccountRequest{
					Account:        fakeAccount(),
					PrivateAccount: fakePrivateAccount(),
					ProjectId:      "test",
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
			Context("SignIn into the account", func() {
				It("should signIn the account and return token and some data", func() {
					signInReq.Username = email
					signInReq.Password = password
					signInReq.Group = group
					signInRes, err := AccountAPI.SignIn(ctx, signInReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(signInRes).ShouldNot(BeNil())
					Expect(signInRes.Token).ShouldNot(BeZero())
					Expect(signInRes.AccountId).ShouldNot(BeZero())
					Expect(signInRes.AccountId).Should(Equal(accountID))
				})
			})
			It("should fail if group is not associated with the account", func() {
				signInReq.Username = email
				signInReq.Password = password
				signInReq.Group = randomdata.Adjective()
				signInRes, err := AccountAPI.SignIn(ctx, signInReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(signInRes).Should(BeNil())
			})
		})

		Describe("SignIn to an account without password", func() {
			var email, group string
			Context("Lets create an account first", func() {
				It("should create account without error", func() {
					createReq := &account.CreateAccountRequest{
						Account:   fakeAccount(),
						ProjectId: "test",
					}
					createRes, err := AccountAPI.CreateAccount(ctx, createReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(createRes).ShouldNot(BeNil())
					email = createReq.Account.Email
					group = createReq.Account.Group
				})
			})
			Context("SignIn into the account", func() {
				It("should fail because password is not present on the account", func() {
					signInReq.Username = email
					signInReq.Password = "hakty11"
					signInReq.Group = group
					signInRes, err := AccountAPI.SignIn(ctx, signInReq)
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.PermissionDenied))
					Expect(signInRes).Should(BeNil())
				})
			})
		})
	})
})
