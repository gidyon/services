package push

import (
	"context"

	"github.com/Pallinder/go-randomdata"
	push "github.com/gidyon/services/pkg/api/messaging/pusher"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func fakePushMessage() *push.PushMessage {
	return &push.PushMessage{
		DeviceTokens: []string{
			randomdata.MacAddress(), randomdata.MacAddress(), randomdata.MacAddress(),
		},
		Title:   randomdata.Adjective(),
		Message: randomdata.Paragraph(),
		Details: map[string]string{
			"mode": "test",
		},
	}
}

var _ = Describe("Sending push", func() {
	var (
		sendReq *push.PushMessage
		ctx     context.Context
	)

	BeforeEach(func() {
		sendReq = fakePushMessage()
		ctx = context.Background()
	})

	Describe("Sending push with malformed request", func() {
		It("should fail when the request is nil", func() {
			sendReq = nil
			sendRes, err := PushServer.SendPushMessage(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail when the device tokens is empty", func() {
			sendReq.DeviceTokens = nil
			sendRes, err := PushServer.SendPushMessage(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail when the title is empty", func() {
			sendReq.Title = ""
			sendRes, err := PushServer.SendPushMessage(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail when the message is empty", func() {
			sendReq.Message = ""
			sendRes, err := PushServer.SendPushMessage(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail details is empty", func() {
			sendReq.Details = nil
			sendRes, err := PushServer.SendPushMessage(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
	})

	Describe("Sending push with valid request", func() {
		It("should succeed", func() {
			sendRes, err := PushServer.SendPushMessage(ctx, sendReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			Expect(sendRes).ShouldNot(BeNil())
		})
	})
})
