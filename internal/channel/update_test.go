package channel

import (
	"context"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/channel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Updating A Channel #update", func() {
	var (
		updateReq *channel.UpdateChannelRequest
		ctx       context.Context
	)

	BeforeEach(func() {
		updateReq = &channel.UpdateChannelRequest{
			OwnerId: randomdata.RandStringRunes(32),
			Channel: &channel.Channel{
				Title:       randomdata.Month(),
				Description: randomdata.Paragraph(),
			},
		}
		ctx = context.Background()
	})

	Describe("Updating a channel with incorrect/missing values", func() {
		It("should fail when the request is emptyx`", func() {
			updateReq = nil
			updateRes, err := ChannelAPI.UpdateChannel(ctx, updateReq)
			Expect(err).Should(HaveOccurred())
			Expect(updateRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when channel is empty", func() {
			updateReq.Channel = nil
			updateRes, err := ChannelAPI.UpdateChannel(ctx, updateReq)
			Expect(err).Should(HaveOccurred())
			Expect(updateRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when channel id is incorrect", func() {
			updateReq.Channel.Id = randomdata.RandStringRunes(32)
			updateRes, err := ChannelAPI.UpdateChannel(ctx, updateReq)
			Expect(err).Should(HaveOccurred())
			Expect(updateRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when channel id is missing", func() {
			updateReq.OwnerId = ""
			updateRes, err := ChannelAPI.UpdateChannel(ctx, updateReq)
			Expect(err).Should(HaveOccurred())
			Expect(updateRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
	})

	Describe("Updating a channel with correct/valid request", func() {
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

			Describe("updating the channel", func() {
				var title string
				BeforeEach(func() {
					updateReq.Channel.Id = channelID
					updateReq.Channel.Title = randomdata.SillyName()
				})

				It("should succeed when the request is valid", func() {
					updateRes, err := ChannelAPI.UpdateChannel(ctx, updateReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(updateRes).ShouldNot(BeNil())
				})
				It("should succeed when owner id is missing", func() {
					updateReq.Channel.OwnerId = ""
					updateRes, err := ChannelAPI.UpdateChannel(ctx, updateReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(updateRes).ShouldNot(BeNil())
				})
				It("should succeed when channel description is missing", func() {
					updateReq.Channel.Description = ""
					updateRes, err := ChannelAPI.UpdateChannel(ctx, updateReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(updateRes).ShouldNot(BeNil())
					title = updateReq.Channel.Title
				})

				Describe("Getting an updated channel", func() {
					It("should reflect updated fields", func() {
						getRes, err := ChannelAPI.GetChannel(ctx, &channel.GetChannelRequest{
							Id: channelID,
						})
						Expect(err).ShouldNot(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.OK))
						Expect(getRes).ShouldNot(BeNil())
						Expect(getRes.Title).Should(Equal(title))
					})
				})
			})
		})
	})
})
