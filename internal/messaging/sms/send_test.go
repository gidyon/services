package sms

import (
	"context"

	"github.com/Pallinder/go-randomdata"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gidyon/services/pkg/api/messaging/sms"
)

func fakeSMS() *sms.SMS {
	return &sms.SMS{
		DestinationPhones: []string{
			randomdata.PhoneNumber(), randomdata.PhoneNumber(), randomdata.PhoneNumber(),
		},
		Keyword: randomdata.Adjective(),
		Message: randomdata.Paragraph(),
	}
}

var _ = Describe("Sending sms", func() {
	var (
		sendReq *sms.SendSMSRequest
		ctx     context.Context
	)

	BeforeEach(func() {
		sendReq = &sms.SendSMSRequest{
			Sms: fakeSMS(),
		}
		ctx = context.Background()
	})

	Describe("Sending sms with malformed request", func() {
		It("should fail when the request is nil", func() {
			sendReq = nil
			sendRes, err := SMSAPI.SendSMS(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail when the destination phones is empty", func() {
			sendReq.Sms.DestinationPhones = nil
			sendRes, err := SMSAPI.SendSMS(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail when the keyword is empty", func() {
			sendReq.Sms.Keyword = ""
			sendRes, err := SMSAPI.SendSMS(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail when the message is empty", func() {
			sendReq.Sms.Message = ""
			sendRes, err := SMSAPI.SendSMS(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
	})

	Describe("Sending sms with valid request", func() {
		It("should succeed", func() {
			sendRes, err := SMSAPI.SendSMS(ctx, sendReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			Expect(sendRes).ShouldNot(BeNil())
		})
	})
})
