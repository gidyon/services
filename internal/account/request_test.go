package account

import (
	"context"

	"github.com/gidyon/services/pkg/api/messaging"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/account"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Requesting for an account reset @reset", func() {
	var (
		resetReq *account.RequestChangePrivateAccountRequest
		ctx      context.Context
	)

	BeforeEach(func() {
		resetReq = &account.RequestChangePrivateAccountRequest{
			Payload:     randomdata.Email(),
			FallbackUrl: randomdata.MacAddress(),
			SendMethod:  messaging.SendMethod_EMAIL,
		}
		ctx = context.Background()
	})

	Describe("Requesting for change token with malformed request", func() {
		It("should fail when the request is nil", func() {
			resetReq = nil
			updateRes, err := AccountAPI.RequestChangePrivateAccount(ctx, resetReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(updateRes).Should(BeNil())
		})
		It("should fail when the payload in request is missing", func() {
			resetReq.Payload = ""
			updateRes, err := AccountAPI.RequestChangePrivateAccount(ctx, resetReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(updateRes).Should(BeNil())
		})
		It("should fail when fallback url in request is missing", func() {
			resetReq.FallbackUrl = ""
			updateRes, err := AccountAPI.RequestChangePrivateAccount(ctx, resetReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(updateRes).Should(BeNil())
		})
		It("should fail when send method is unknown", func() {
			resetReq.SendMethod = messaging.SendMethod_SEND_METHOD_UNSPECIFIED
			updateRes, err := AccountAPI.RequestChangePrivateAccount(ctx, resetReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(updateRes).Should(BeNil())
		})
	})

	Describe("Requesting for change token with well-formed request", func() {
		It("should fail when the account does not exist", func() {
			updateRes, err := AccountAPI.RequestChangePrivateAccount(ctx, resetReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.NotFound))
			Expect(updateRes).Should(BeNil())
		})

		Context("Requesting for token with an existing account", func() {
			var payload string
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
					payload = createReq.Account.Email
				})
			})

			It("should succeed in requesting the token", func() {
				resetReq.Payload = payload
				updateRes, err := AccountAPI.RequestChangePrivateAccount(ctx, resetReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(updateRes).ShouldNot(BeNil())
			})
		})
	})
})
