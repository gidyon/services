package messaging

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/messaging"
)

var _ = Describe("Broadcasting a message to many users @broadcast", func() {
	var (
		broadCastReq *messaging.BroadCastMessageRequest
		ctx          context.Context
		userID       = fmt.Sprint(randomdata.Number(100, 999))
	)

	BeforeEach(func() {
		broadCastReq = &messaging.BroadCastMessageRequest{
			Message:  fakeMessage(userID),
			Channels: []string{"default"},
		}
		ctx = context.Background()
	})

	Describe("Broadcasting message with mallformed request", func() {
		It("should fail when the request is nil", func() {
			broadCastReq = nil
			sendRes, err := MessagingAPI.BroadCastMessage(ctx, broadCastReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail if message is nil", func() {
			broadCastReq.Message = nil
			sendRes, err := MessagingAPI.BroadCastMessage(ctx, broadCastReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail if title is missing", func() {
			broadCastReq.Message.Title = ""
			sendRes, err := MessagingAPI.BroadCastMessage(ctx, broadCastReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail if message data is missing", func() {
			broadCastReq.Message.Data = ""
			sendRes, err := MessagingAPI.BroadCastMessage(ctx, broadCastReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail if channels is missing", func() {
			broadCastReq.Channels = nil
			sendRes, err := MessagingAPI.BroadCastMessage(ctx, broadCastReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
	})

	Describe("Broadcasting message with a well formed request", func() {
		It("should succeed in broadcasting user message", func() {
			broadCastRes, err := MessagingAPI.BroadCastMessage(ctx, broadCastReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			Expect(broadCastRes).ShouldNot(BeNil())
		})

		for _, sendMethod := range messaging.SendMethod_value {
			sendMethod := sendMethod
			Describe("Broadcasting message with different send methods", func() {
				BeforeEach(func() {
					broadCastReq = &messaging.BroadCastMessageRequest{
						Message:  fakeMessage(userID),
						Channels: []string{"default"},
					}
				})

				// Otherwise
				It("should succeed", func() {
					broadCastReq.Message.SendMethods = []messaging.SendMethod{messaging.SendMethod(sendMethod)}
					sendRes, err := MessagingAPI.BroadCastMessage(ctx, broadCastReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(sendRes).ShouldNot(BeNil())
					Expect(err).ShouldNot(HaveOccurred())
				})
			})
		}
	})
})
