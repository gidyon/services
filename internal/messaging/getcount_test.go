package messaging

import (
	"context"
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/messaging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Getting how many new messages @count", func() {
	var (
		userID = fmt.Sprint(randomdata.Number(100, 999))
		getReq *messaging.MessageRequest
		ctx    context.Context
	)

	BeforeEach(func() {
		getReq = &messaging.MessageRequest{
			UserId: userID,
		}
		ctx = context.Background()
	})

	Describe("Getting how many new messages with malformed request", func() {
		It("should fail when the request is nil", func() {
			getReq = nil
			getRes, err := MessagingAPI.GetNewMessagesCount(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(getRes).Should(BeNil())
		})
		It("should fail when user id is missing in request", func() {
			getReq.UserId = ""
			getRes, err := MessagingAPI.GetNewMessagesCount(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(getRes).Should(BeNil())
		})
		It("should fail when user id is incorrect", func() {
			getReq.UserId = randomdata.RandStringRunes(32)
			getRes, err := MessagingAPI.GetNewMessagesCount(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(getRes).Should(BeNil())
		})
	})

	Describe("Getting how many new messages with correct request", func() {
		Context("Lets create a message first", func() {
			It("should succeed in creating a message", func() {
				messagePB := fakeMessage(userID)
				_, err := MessagingAPI.SendMessage(ctx, &messaging.SendMessageRequest{
					Message: messagePB,
				})
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Describe("Getting how many new messages should succeed", func() {
			It("should succeed", func() {
				getRes, err := MessagingAPI.GetNewMessagesCount(ctx, getReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(getRes).ShouldNot(BeNil())
				Expect(getRes.Count).ShouldNot(BeZero())
			})
		})
	})
})
