package subscriber

import (
	"context"
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/subscriber"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Subsribing A User To A Channel @subscribe", func() {
	var subscriberID = fmt.Sprint(randomdata.Number(1000000, 1999999))

	var (
		subscribeReq *subscriber.SubscriberRequest
		ctx          context.Context
	)

	BeforeEach(func() {
		subscribeReq = &subscriber.SubscriberRequest{
			SubscriberId: uuid.New().String(),
			ChannelName:  randomdata.Month(),
			ChannelId:    randomdata.RandStringRunes(32),
		}
		ctx = context.Background()
	})

	Describe("Subscribing to a channel with malformed request", func() {
		It("should fail when the request is nil", func() {
			subscribeReq = nil
			subscribeRes, err := SubsriberAPI.Subscribe(ctx, subscribeReq)
			Expect(err).Should(HaveOccurred())
			Expect(subscribeRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when subscriber id is missing", func() {
			subscribeReq.SubscriberId = ""
			subscribeRes, err := SubsriberAPI.Subscribe(ctx, subscribeReq)
			Expect(err).Should(HaveOccurred())
			Expect(subscribeRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when channel id is missing", func() {
			subscribeReq.ChannelId = ""
			subscribeRes, err := SubsriberAPI.Subscribe(ctx, subscribeReq)
			Expect(err).Should(HaveOccurred())
			Expect(subscribeRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when channel name is missing", func() {
			subscribeReq.ChannelName = ""
			subscribeRes, err := SubsriberAPI.Subscribe(ctx, subscribeReq)
			Expect(err).Should(HaveOccurred())
			Expect(subscribeRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when subscriber is is incorrect", func() {
			subscribeReq.ChannelName = uuid.New().String()
			subscribeRes, err := SubsriberAPI.Subscribe(ctx, subscribeReq)
			Expect(err).Should(HaveOccurred())
			Expect(subscribeRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
	})

	Describe("Subscribing to a channel with well-formed request", func() {
		var channelID string
		It("should subscribe user when the request is valid", func() {
			subscribeReq.SubscriberId = subscriberID
			subscribeRes, err := SubsriberAPI.Subscribe(ctx, subscribeReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			Expect(subscribeRes).ShouldNot(BeNil())
			channelID = subscribeReq.ChannelId
		})

		Describe("Getting the susbscriber", func() {
			It("should succeed and susbcribed channel present", func() {
				subscriberPB, err := SubsriberAPI.GetSubscriber(ctx, &subscriber.GetSubscriberRequest{
					SubscriberId: subscriberID,
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(subscriberPB).ShouldNot(BeNil())
				var exist bool
				for _, channelPB := range subscriberPB.Channels {
					if channelPB.ChannelId == channelID {
						exist = true
					}
				}
				Expect(exist).Should(BeTrue())
			})
		})
	})
})
