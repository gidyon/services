package channel

import (
	"context"

	"github.com/Pallinder/go-randomdata"

	"github.com/gidyon/services/pkg/api/channel"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Deleting A Channel #delete", func() {
	var (
		deleteReq *channel.DeleteChannelRequest
		ctx       context.Context
	)

	BeforeEach(func() {
		deleteReq = &channel.DeleteChannelRequest{
			Id:      uuid.New().String(),
			OwnerId: randomdata.RandStringRunes(32),
		}
		ctx = context.Background()
	})

	Describe("Deleting a channel with incorrect/missing values", func() {
		It("should fail when the request is nil", func() {
			deleteReq = nil
			deleteRes, err := ChannelAPI.DeleteChannel(ctx, deleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(deleteRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when channel id is missing", func() {
			deleteReq.Id = ""
			deleteRes, err := ChannelAPI.DeleteChannel(ctx, deleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(deleteRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when channel id is incorrect", func() {
			deleteReq.Id = randomdata.RandStringRunes(32)
			deleteRes, err := ChannelAPI.DeleteChannel(ctx, deleteReq)
			Expect(err).Should(HaveOccurred())
			Expect(deleteRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
	})

	Describe("Deleting a channel with correct/valid request", func() {
		var channelID string
		Describe("Lets create the channel first", func() {
			It("should succeed", func() {
				createReq := &channel.CreateChannelRequest{
					Channel: mockChannel(),
				}
				createRes, err := ChannelAPI.CreateChannel(ctx, createReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(createRes).ShouldNot(BeNil())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				channelID = createRes.Id
			})
			Describe("Deleting the channel", func() {
				It("should succeed when the request is valid", func() {
					deleteReq.Id = channelID
					deleteRes, err := ChannelAPI.DeleteChannel(ctx, deleteReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(deleteRes).ShouldNot(BeNil())
					Expect(status.Code(err)).Should(Equal(codes.OK))
				})
			})

			Describe("Getting the deleted channel", func() {
				It("should not exist in database", func() {
					getRes, err := ChannelAPI.GetChannel(ctx, &channel.GetChannelRequest{
						Id: channelID,
					})
					Expect(err).Should(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.NotFound))
					Expect(getRes).Should(BeNil())
				})
			})
		})

	})
})
