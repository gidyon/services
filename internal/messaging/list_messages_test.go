package messaging

import (
	"context"
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/messaging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Getting messages @list", func() {
	var (
		listReq *messaging.ListMessagesRequest
		ctx     context.Context
		userID  = fmt.Sprint(randomdata.Number(100, 999))
	)

	BeforeEach(func() {
		listReq = &messaging.ListMessagesRequest{
			Filter: &messaging.ListMessagesFilter{
				UserId: userID,
			},
		}
		ctx = context.Background()
	})

	Describe("Getting messages with malformed request", func() {
		It("should fail when the request is nil", func() {
			listReq = nil
			getRes, err := MessagingAPI.ListMessages(ctx, listReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(getRes).Should(BeNil())
		})
		It("should fail when user id has incorrect syntax", func() {
			listReq.Filter.UserId = randomdata.RandStringRunes(32)
			getRes, err := MessagingAPI.ListMessages(ctx, listReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(getRes).Should(BeNil())
		})
		It("should fail when page token is incorrect", func() {
			listReq.PageToken = randomdata.RandStringRunes(32)
			getRes, err := MessagingAPI.ListMessages(ctx, listReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(getRes).Should(BeNil())
		})
	})

	Describe("Getting messages with correct request", func() {
		Context("Lets create a message first", func() {
			It("should succed in creating a message", func() {
				messagePB := fakeMessage(userID)
				sendRes, err := MessagingAPI.SendMessage(ctx, messagePB)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(sendRes).ShouldNot(BeNil())
			})
		})

		When("Getting messages it should succeed", func() {
			It("should succeed", func() {
				getRes, err := MessagingAPI.ListMessages(ctx, listReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(getRes.Messages).ShouldNot(BeNil())
				Expect(len(getRes.Messages)).ShouldNot(BeZero())
			})
		})

		When("Getting messages without specifying user id should succeed", func() {
			It("should succeed", func() {
				getRes, err := MessagingAPI.ListMessages(ctx, listReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(getRes.Messages).ShouldNot(BeNil())
				Expect(len(getRes.Messages)).ShouldNot(BeZero())
			})
		})

		When("Getting messages with type all should succeed", func() {
			It("should succeed", func() {
				listReq.Filter.UserId = ""
				listReq.Filter.TypeFilters = []messaging.MessageType{
					messaging.MessageType_ALL,
				}
				getRes, err := MessagingAPI.ListMessages(ctx, listReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(getRes.Messages).ShouldNot(BeNil())
				Expect(len(getRes.Messages)).ShouldNot(BeZero())
			})
		})

		When("Getting messages with type filters should succeed", func() {
			It("should succeed", func() {
				listReq.Filter.UserId = ""
				listReq.Filter.TypeFilters = []messaging.MessageType{
					messaging.MessageType_PROMOTIONAL,
					messaging.MessageType_REMINDER,
					messaging.MessageType_WARNING,
				}
				getRes, err := MessagingAPI.ListMessages(ctx, listReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(getRes.Messages).ShouldNot(BeNil())
				Expect(len(getRes.Messages)).ShouldNot(BeZero())
				for _, messagePB := range getRes.Messages {
					Expect(messagePB.Type).Should(BeElementOf(listReq.Filter.TypeFilters))
				}
			})
		})
	})
})
