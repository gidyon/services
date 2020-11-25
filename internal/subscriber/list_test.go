package subscriber

import (
	"context"
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/subscriber"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Listing Subscribers For A Channel @list", func() {

	var subscriberID = fmt.Sprint(randomdata.Number(1000000, 1999999))

	var (
		listReq *subscriber.ListSubscribersRequest
		ctx     context.Context
	)

	BeforeEach(func() {
		listReq = &subscriber.ListSubscribersRequest{
			PageToken: "",
			PageSize:  10,
			Filter: &subscriber.ListSubscribersFilter{
				Channels: []string{"public"},
			},
		}
		ctx = context.Background()
	})

	Describe("Listing subscribers from a channel with malformed request", func() {
		It("should fail when the request is nil", func() {
			listReq = nil
			listRes, err := SubsriberAPI.ListSubscribers(ctx, listReq)
			Expect(err).Should(HaveOccurred())
			Expect(listRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when page token is incorrect", func() {
			listReq.PageToken = randomdata.RandStringRunes(64)
			listRes, err := SubsriberAPI.ListSubscribers(ctx, listReq)
			Expect(err).Should(HaveOccurred())
			Expect(listRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
	})

	Describe("Listing subscribers for a channel with valid request", func() {
		var (
			channelName = randomdata.Month()
			channelID   = randomdata.RandStringRunes(32)
			pageToken   string
		)

		Describe("List subscribers when missing channelId", func() {
			It("should list subscribers for a channel even when page ", func() {
				listReq.Filter.Channels = []string{}
				listReq.PageToken = pageToken
				listRes, err := SubsriberAPI.ListSubscribers(ctx, listReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(listRes).ShouldNot(BeNil())
				Expect(len(listRes.Subscribers)).ShouldNot(BeZero())
			})
		})

		Context("Lets subscribe an account to one channel", func() {
			It("should subscribe account to channel", func() {
				subscribeRes, err := SubsriberAPI.Subscribe(ctx, &subscriber.SubscriberRequest{
					SubscriberId: subscriberID,
					ChannelId:    channelID,
					ChannelName:  channelName,
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(subscribeRes).ShouldNot(BeNil())
			})

			It("should list subscribers for a channel", func() {
				listReq.Filter.Channels = []string{channelID}
				listRes, err := SubsriberAPI.ListSubscribers(ctx, listReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(listRes).ShouldNot(BeNil())
				for _, subscriberPB := range listRes.Subscribers {
					Expect(channelID).Should(BeElementOf(getChannelNames(subscriberPB.Channels)))
				}
				pageToken = listRes.NextPageToken
			})

			Describe("Listing channels using previous next_page_token as page_token", func() {
				It("should list subscribers for a channel even when page ", func() {
					listReq.Filter.Channels = []string{channelID}
					listReq.PageToken = pageToken
					listRes, err := SubsriberAPI.ListSubscribers(ctx, listReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(listRes).ShouldNot(BeNil())
				})
			})
		})
	})
})

func getChannelNames(channelsPB []*subscriber.ChannelSubcriber) []string {
	channels := make([]string, 0, len(channelsPB))
	for _, channelPB := range channelsPB {
		channels = append(channels, channelPB.ChannelId)
	}
	return channels
}
