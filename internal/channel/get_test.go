package channel

import (
	"context"
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/channel"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("GetChannel For A User #get", func() {
	var (
		getReq *channel.GetChannelRequest
		ctx    context.Context
	)

	BeforeEach(func() {
		getReq = &channel.GetChannelRequest{
			Id: uuid.New().String(),
		}
		ctx = context.Background()
	})

	Describe("Calling GetChannel with missing/incorrect values", func() {
		It("should fail when the request is nil", func() {
			getReq = nil
			getRes, err := ChannelAPI.GetChannel(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(getRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when channel id is missing", func() {
			getReq.Id = ""
			getRes, err := ChannelAPI.GetChannel(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(getRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when channel id is incorrect", func() {
			getReq.Id = "i dont exist"
			getRes, err := ChannelAPI.GetChannel(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(getRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when channel does not exist", func() {
			getReq.Id = fmt.Sprint(randomdata.Number(1000099999, 9999191999))
			getRes, err := ChannelAPI.GetChannel(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(getRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.NotFound))
		})
	})

	Describe("Calling GetChannel with correct request", func() {
		var channelID string
		Context("Let's create a channel first", func() {
			It("should succeed in creating a channel", func() {
				createReq := &channel.CreateChannelRequest{
					Channel: mockChannel(),
				}
				createRes, err := ChannelAPI.CreateChannel(ctx, createReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(createRes).ShouldNot(BeNil())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				channelID = createRes.Id
			})
		})

		When("Getting the channel", func() {
			It("should succeed", func() {
				getReq.Id = channelID
				getRes, err := ChannelAPI.GetChannel(ctx, getReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(getRes).ShouldNot(BeNil())
			})
		})
	})
})
