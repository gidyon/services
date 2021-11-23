package account

import (
	"context"
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/account"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Creating Account @create", func() {

	var (
		createReq *account.CreateAccountRequest
		ctx       context.Context
	)

	BeforeEach(func() {
		createReq = &account.CreateAccountRequest{
			Account:        fakeAccount(),
			PrivateAccount: fakePrivateAccount(),
			ProjectId:      "1",
		}
		ctx = context.Background()
	})

	Context("Failing Scenarios", func() {
		When("Creating account with nil request", func() {
			It("should fail when request is nil", func() {
				createReq = nil
				createRes, err := AccountAPI.CreateAccount(ctx, createReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(createRes).Should(BeNil())
			})
		})

		When("Creating account with some missing request fields", func() {
			It("should fail when projevt id is missing", func() {
				createReq.ProjectId = ""
				createRes, err := AccountAPI.CreateAccount(ctx, createReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(createRes).Should(BeNil())
			})
			It("should fail when names is missing", func() {
				createReq.Account.Names = ""
				createRes, err := AccountAPI.CreateAccount(ctx, createReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(createRes).Should(BeNil())
			})
			It("should fail when account group is missing", func() {
				createReq.Account.Group = ""
				createRes, err := AccountAPI.CreateAccount(ctx, createReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(createRes).Should(BeNil())
			})
			It("should fail when both phone and email id are missing", func() {
				createReq.Account.Phone = ""
				createReq.Account.Email = ""
				createRes, err := AccountAPI.CreateAccount(ctx, createReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(createRes).Should(BeNil())
			})
			It("should fail when by admin is true and admin id is missing", func() {
				createReq.ByAdmin = true
				createReq.AdminId = ""
				createRes, err := AccountAPI.CreateAccount(ctx, createReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(createRes).Should(BeNil())
			})
		})

		When("Creating account with email or phone that already exists in database", func() {
			var email, phone string
			Context("Lets create an account", func() {
				It("should create account in database without error", func() {
					createRes, err := AccountAPI.CreateAccount(ctx, createReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(createRes).ShouldNot(BeNil())
					email = createReq.Account.Email
					phone = createReq.Account.Phone
				})

				Describe("Creating account with existing email or phone", func() {
					It("should fail when email is already registered", func() {
						createReq.Account.Email = email
						createRes, err := AccountAPI.CreateAccount(ctx, createReq)
						Expect(err).Should(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.AlreadyExists))
						Expect(createRes).Should(BeNil())
					})
					It("should fail when phone is already registered", func() {
						createReq.Account.Phone = phone
						createRes, err := AccountAPI.CreateAccount(ctx, createReq)
						Expect(err).Should(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.AlreadyExists))
						Expect(createRes).Should(BeNil())
					})
				})
			})
		})
	})

	Context("Success scenarios", func() {
		When("Creating account with valid request", func() {
			It("should succeed when email is missing but phone is not", func() {
				createReq.Account.Email = ""
				createReq.Account.Phone = randomdata.PhoneNumber()[:10]
				createRes, err := AccountAPI.CreateAccount(ctx, createReq)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(createRes).ShouldNot(BeNil())
			})

			It("should succeed when phone is missing but email is not", func() {
				createReq.Account.Phone = ""
				createReq.Account.Email = randomdata.Email()
				createRes, err := AccountAPI.CreateAccount(ctx, createReq)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(createRes).ShouldNot(BeNil())
			})

			When("Creating account when all request fields provided", func() {
				It("should create account in database without error", func() {
					createRes, err := AccountAPI.CreateAccount(ctx, createReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(createRes).ShouldNot(BeNil())
				})
			})
		})

		When("Creating an account by admin", func() {
			It("should create account in database without error", func() {
				createReq.ByAdmin = true
				createReq.AdminId = fmt.Sprint(randomdata.Decimal(10))
				createRes, err := AccountAPI.CreateAccount(ctx, createReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(createRes).ShouldNot(BeNil())
			})
		})
	})
})
