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

var _ = Describe("Unsubscribing A User From A Channel @unsubscribe", func() {

	var subscriberID = fmt.Sprint(randomdata.Number(1000000, 1999999))

	var (
		subscribeReq *subscriber.SubscriberRequest
		ctx          context.Context
	)

	BeforeEach(func() {
		subscribeReq = &subscriber.SubscriberRequest{
			SubscriberId: uuid.New().String(),
			Channels:     []string{randomdata.Month()},
		}
		ctx = context.Background()
	})

	Describe("Unsubscribing from a channel with malformed request", func() {
		It("should fail when the request is nil", func() {
			subscribeReq = nil
			subscribeRes, err := SubsriberAPI.Unsubscribe(ctx, subscribeReq)
			Expect(err).Should(HaveOccurred())
			Expect(subscribeRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when subscriber id is is missing", func() {
			subscribeReq.SubscriberId = ""
			subscribeRes, err := SubsriberAPI.Unsubscribe(ctx, subscribeReq)
			Expect(err).Should(HaveOccurred())
			Expect(subscribeRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when channel names is missing", func() {
			subscribeReq.Channels = nil
			subscribeRes, err := SubsriberAPI.Unsubscribe(ctx, subscribeReq)
			Expect(err).Should(HaveOccurred())
			Expect(subscribeRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
	})

	Describe("Unsubscribing from a channel with well-formed request", func() {
		var channelName string

		Describe("Lets subscribe user to a channel first", func() {
			It("should subscribe user when the request is valid", func() {
				subscribeReq.SubscriberId = subscriberID
				subscribeRes, err := SubsriberAPI.Subscribe(ctx, subscribeReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(subscribeRes).ShouldNot(BeNil())
				channelName = subscribeReq.Channels[0]
			})

			Describe("Getting the subscriber", func() {
				It("should succeed and subscribed channel present", func() {
					subscriberPB, err := SubsriberAPI.GetSubscriber(ctx, &subscriber.GetSubscriberRequest{
						SubscriberId: subscriberID,
					})
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(subscriberPB).ShouldNot(BeNil())
					var exist bool
					for _, ch := range subscriberPB.Channels {
						if ch == channelName {
							exist = true
						}
					}
					Expect(exist).Should(BeTrue())
				})
			})

			Describe("Unsubscribing user from the channel", func() {
				It("should unsubscribe user from channel", func() {
					subscribeReq.SubscriberId = subscriberID
					subscribeReq.Channels = []string{channelName}
					subscribeRes, err := SubsriberAPI.Unsubscribe(ctx, subscribeReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(subscribeRes).ShouldNot(BeNil())
				})

				Describe("Getting the subscriber", func() {
					It("should succeed and subscribed channel absent", func() {
						subscriberPB, err := SubsriberAPI.GetSubscriber(ctx, &subscriber.GetSubscriberRequest{
							SubscriberId: subscriberID,
						})
						Expect(err).ShouldNot(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.OK))
						Expect(subscriberPB).ShouldNot(BeNil())
						var exist bool
						for _, ch := range subscriberPB.Channels {
							if ch == channelName {
								exist = true
							}
						}
						Expect(exist).Should(BeFalse())
					})
				})
			})
		})

		Describe("Unsubscribing user who is not subscribed to a channel", func() {
			var subscriberID = fmt.Sprint(randomdata.Number(1000000, 1999999))

			It("should create a susbcriber with default channel", func() {
				subscribeReq.SubscriberId = subscriberID
				subscribeReq.Channels = []string{channelName}
				subscribeRes, err := SubsriberAPI.Unsubscribe(ctx, subscribeReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(subscribeRes).ShouldNot(BeNil())
			})

			Describe("Getting the subscriber", func() {
				It("should succeed", func() {
					subscriberPB, err := SubsriberAPI.GetSubscriber(ctx, &subscriber.GetSubscriberRequest{
						SubscriberId: subscriberID,
					})
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(subscriberPB).ShouldNot(BeNil())
				})
			})
		})
	})
})
