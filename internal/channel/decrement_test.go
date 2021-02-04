package channel

import (
	"context"

	"github.com/Pallinder/go-randomdata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gidyon/services/pkg/api/channel"
)

var _ = Describe("Decrementing subscribers for a channel @decrement", func() {
	var (
		decrReq *channel.SubscribersRequest
		ctx     context.Context
	)

	BeforeEach(func() {
		decrReq = &channel.SubscribersRequest{
			ChannelName: randomdata.Adjective(),
		}
		ctx = context.Background()
	})

	Describe("Decrementing with malformed request", func() {
		It("should fail when the request is nil", func() {
			decrReq = nil
			decrRes, err := ChannelAPI.DecrementSubscribers(ctx, decrReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(decrRes).Should(BeNil())
		})
		It("should fail when channel name is missing", func() {
			decrReq.ChannelName = ""
			decrRes, err := ChannelAPI.DecrementSubscribers(ctx, decrReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(decrRes).Should(BeNil())
		})
	})

	Describe("Decrementing channel with well-formed request", func() {
		var channelName, channelID string
		Context("Lets create one channel first", func() {
			It("should succeed", func() {
				createReq := &channel.CreateChannelRequest{
					Channel: mockChannel(),
				}
				createRes, err := ChannelAPI.CreateChannel(ctx, createReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(createRes).ShouldNot(BeNil())
				channelName = createReq.Channel.Title
				channelID = createRes.Id
			})
		})

		count := 20

		When("Getting the channel", func() {
			It("should succeed", func() {
				getRes, err := ChannelAPI.GetChannel(ctx, &channel.GetChannelRequest{
					Id: channelID,
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(getRes).ShouldNot(BeNil())
			})

			// Lets decrrement
			for i := 0; i < count; i++ {
				It("should succeed", func() {
					decrReq.ChannelName = channelName
					incRes, err := ChannelAPI.IncrementSubscribers(ctx, decrReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(incRes).ShouldNot(BeNil())
				})
			}

			// Lets decrement
			for i := 0; i < count/2; i++ {
				It("should succeed", func() {
					decrReq.ChannelName = channelName
					decrRes, err := ChannelAPI.DecrementSubscribers(ctx, decrReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(decrRes).ShouldNot(BeNil())
				})
			}

			When("Getting the channel", func() {
				It("should succeed", func() {
					getRes, err := ChannelAPI.GetChannel(ctx, &channel.GetChannelRequest{
						Id: channelID,
					})
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(getRes).ShouldNot(BeNil())
					Expect(getRes.Subscribers).Should(BeEquivalentTo(count / 2))
				})
			})
		})
	})
})
