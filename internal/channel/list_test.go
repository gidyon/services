package channel

import (
	"context"

	"github.com/gidyon/services/pkg/api/channel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("ListChannels #list", func() {
	var (
		listReq *channel.ListChannelsRequest
		ctx     context.Context
	)

	BeforeEach(func() {
		listReq = &channel.ListChannelsRequest{
			PageToken: "",
		}
		ctx = context.Background()
	})

	Describe("Calling ListChannels with missing/incorrect values", func() {
		It("should fail when the request is nil", func() {
			listReq = nil
			listRes, err := ChannelAPI.ListChannels(ctx, listReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(listRes).Should(BeNil())
		})
		It("should fail when page token is icorrect", func() {
			listReq.PageToken = "nil"
			listRes, err := ChannelAPI.ListChannels(ctx, listReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(listRes).Should(BeNil())
		})
	})

	Describe("Calling ListChannels with valid values", func() {
		Context("Lets create one channel first", func() {
			It("should succeed", func() {
				createReq := &channel.CreateChannelRequest{
					Channel: mockChannel(),
				}
				createRes, err := ChannelAPI.CreateChannel(ctx, createReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(createRes).ShouldNot(BeNil())
			})
		})

		It("should succeed when the request is valid", func() {
			listRes, err := ChannelAPI.ListChannels(ctx, listReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			Expect(listRes).ShouldNot(BeNil())
		})

		It("should succeed even when the page token is too big", func() {
			listReq.PageSize = 10
			listRes, err := ChannelAPI.ListChannels(ctx, listReq)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.OK))
			Expect(listRes).ShouldNot(BeNil())
		})
	})
})
