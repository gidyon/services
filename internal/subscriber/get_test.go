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

var _ = Describe("Getting subscriber @get", func() {

	var subscriberID = fmt.Sprint(randomdata.Number(1000000, 1999999))

	var (
		getReq *subscriber.GetSubscriberRequest
		ctx    context.Context
	)

	BeforeEach(func() {
		getReq = &subscriber.GetSubscriberRequest{
			SubscriberId: uuid.New().String(),
		}
		ctx = context.Background()
	})

	Describe("Getting subscriber with malformed request", func() {
		It("should fail when the request is nil", func() {
			getReq = nil
			getRes, err := SubsriberAPI.GetSubscriber(ctx, getReq)
			Expect(err).To(HaveOccurred())
			Expect(getRes).To(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
		It("should fail when the subscriber id is missing in request", func() {
			getReq.SubscriberId = ""
			getRes, err := SubsriberAPI.GetSubscriber(ctx, getReq)
			Expect(err).To(HaveOccurred())
			Expect(getRes).To(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})

		It("should fail when subscriber id is incorrect", func() {
			getReq.SubscriberId = uuid.New().String()
			subscribeRes, err := SubsriberAPI.GetSubscriber(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(subscribeRes).Should(BeNil())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})
	})

	Describe("Getting subscriber with valid request", func() {
		channelName := randomdata.Month()
		Context("Lets create subscriber account first", func() {
			It("should create subscriber account", func() {
				subscribeRes, err := SubsriberAPI.Subscribe(ctx, &subscriber.SubscriberRequest{
					SubscriberId: subscriberID,
					ChannelId:    randomdata.RandStringRunes(32),
					ChannelName:  channelName,
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(subscribeRes).ShouldNot(BeNil())
			})
		})

		Describe("Getting the subscriber", func() {
			It("should succeed when account id exists and is valid", func() {
				getReq.SubscriberId = subscriberID
				getRes, err := SubsriberAPI.GetSubscriber(ctx, getReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(getRes).ShouldNot(BeNil())
			})
		})

		Describe("Getting subscriber who does not exist yet", func() {
			It("should create the subscriber and return them", func() {
				getReq.SubscriberId = fmt.Sprint(randomdata.Number(1000000, 1999999))
				getRes, err := SubsriberAPI.GetSubscriber(ctx, getReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(getRes).ToNot(BeNil())
			})
		})
	})
})
