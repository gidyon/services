package emailing

import (
	"context"

	"github.com/Pallinder/go-randomdata"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gidyon/services/pkg/api/messaging/emailing"
)

func fakeEmail() *emailing.Email {
	return &emailing.Email{
		Destinations: []string{
			randomdata.Email(), randomdata.Email(), randomdata.Email(),
		},
		From:    randomdata.Email(),
		Subject: randomdata.Adjective(),
		Body:    randomdata.Paragraph(),
	}
}

var _ = Describe("Sending email message", func() {
	var (
		sendReq *emailing.SendEmailRequest
		ctx     context.Context
	)

	BeforeEach(func() {
		sendReq = &emailing.SendEmailRequest{
			Email: fakeEmail(),
		}
		ctx = context.Background()
	})

	Describe("Sending email with malformed request", func() {
		It("should fail when request is nil", func() {
			sendReq = nil
			sendRes, err := EmailAPI.SendEmail(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail when email is nil", func() {
			sendReq.Email = nil
			sendRes, err := EmailAPI.SendEmail(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail when destinations is empty", func() {
			sendReq.Email.Destinations = nil
			sendRes, err := EmailAPI.SendEmail(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail when subject is empty", func() {
			sendReq.Email.Subject = ""
			sendRes, err := EmailAPI.SendEmail(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail when body is empty", func() {
			sendReq.Email.Body = ""
			sendRes, err := EmailAPI.SendEmail(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail when from is empty", func() {
			sendReq.Email.From = ""
			sendRes, err := EmailAPI.SendEmail(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
	})

	Describe("Sending emailing with valid request", func() {
		It("should succeed", func() {
			sendRes, err := EmailAPI.SendEmail(ctx, sendReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			Expect(sendRes).ShouldNot(BeNil())
		})
	})
})
