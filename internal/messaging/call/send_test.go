package call

import (
	"context"

	"github.com/Pallinder/go-randomdata"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gidyon/services/pkg/api/messaging/call"
)

func fakeCall() *call.CallPayload {
	return &call.CallPayload{
		DestinationPhones: []string{
			randomdata.PhoneNumber(), randomdata.PhoneNumber(), randomdata.PhoneNumber(),
		},
		Keyword: randomdata.Adjective(),
		Message: randomdata.Paragraph(),
	}
}

var _ = Describe("Sending call", func() {
	var (
		sendReq *call.CallPayload
		ctx     context.Context
	)

	BeforeEach(func() {
		sendReq = fakeCall()
		ctx = context.Background()
	})

	Describe("Sending call with malformed request", func() {
		It("should fail when the request is nil", func() {
			sendReq = nil
			sendRes, err := CallAPI.Call(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail when the destination phones is empty", func() {
			sendReq.DestinationPhones = nil
			sendRes, err := CallAPI.Call(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail when the keyword is empty", func() {
			sendReq.Keyword = ""
			sendRes, err := CallAPI.Call(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail when the message is empty", func() {
			sendReq.Message = ""
			sendRes, err := CallAPI.Call(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
	})

	Describe("Sending call with valid request", func() {
		It("should succeed", func() {
			sendRes, err := CallAPI.Call(ctx, sendReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			Expect(sendRes).ShouldNot(BeNil())
		})
	})
})
