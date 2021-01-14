package account

import (
	"context"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/account"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Checking whether Phone or Email @exist", func() {
	var (
		existReq *account.ExistAccountRequest
		ctx      context.Context
	)

	BeforeEach(func() {
		existReq = &account.ExistAccountRequest{
			Email:     randomdata.Email(),
			Phone:     randomdata.PhoneNumber(),
			ProjectId: "1",
		}
		ctx = context.Background()
	})

	Context("Failure scenarios", func() {
		When("Checking if email or phone exists with bad credentials", func() {
			It("should fail when the request is nil", func() {
				existReq = nil
				existRes, err := AccountAPI.ExistAccount(ctx, existReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(existRes).Should(BeNil())
			})

			It("should fail when email and phone are missing", func() {
				existReq.Email = ""
				existReq.Phone = ""
				existRes, err := AccountAPI.ExistAccount(ctx, existReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(existRes).Should(BeNil())
			})
			It("should fail when project id missing", func() {
				existReq.ProjectId = ""
				existRes, err := AccountAPI.ExistAccount(ctx, existReq)
				Expect(err).Should(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(existRes).Should(BeNil())
			})
		})
	})

	Context("Success scenarios", func() {
		When("Checking existence with correct credentials", func() {
			It("should succeed when only email is given", func() {
				existReq.Phone = ""
				existRes, err := AccountAPI.ExistAccount(ctx, existReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(existRes).ShouldNot(BeNil())
			})
			It("should succeed when only phone is given", func() {
				existReq.Email = ""
				existRes, err := AccountAPI.ExistAccount(ctx, existReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(existRes).ShouldNot(BeNil())
			})
			It("should succeed with when email and phone and project id are given", func() {
				existRes, err := AccountAPI.ExistAccount(ctx, existReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(existRes).ShouldNot(BeNil())
			})
		})

		Describe("Creating an account then checking if it exists", func() {
			var phone, email string
			It("should succeed in creating an account", func() {
				createReq := &account.CreateAccountRequest{
					Account:        fakeAccount(),
					PrivateAccount: fakePrivateAccount(),
					ProjectId:      "1",
				}
				createRes, err := AccountAPI.CreateAccount(ctx, createReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(createRes).ShouldNot(BeNil())
				phone = createReq.Account.Phone
				email = createReq.Account.Email
			})

			When("Checking existence with credentials that already exist", func() {
				It("should succeed if phone already exists and existence should be true", func() {
					existReq.Phone = phone
					existRes, err := AccountAPI.ExistAccount(ctx, existReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(existRes).ShouldNot(BeNil())
					Expect(existRes.Exists).Should(BeTrue())
				})
				It("should succeed if email already exists and existence should be true", func() {
					existReq.Email = email
					existRes, err := AccountAPI.ExistAccount(ctx, existReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(existRes).ShouldNot(BeNil())
					Expect(existRes.Exists).Should(BeTrue())
				})
			})
		})
	})
})
