package messaging

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Pallinder/go-randomdata"

	"github.com/gidyon/services/pkg/api/messaging"
)

func randomType() messaging.MessageType {
	return messaging.MessageType(rand.Intn(len(messaging.MessageType_name) + 1))
}

func randomSendMethod() messaging.SendMethod {
	index := rand.Intn(len(messaging.SendMethod_name) + 1)
	if index == 0 {
		index = 1
	}
	return messaging.SendMethod(index)
}

func randomPhone() string {
	return randomdata.PhoneNumber()[:12]
}

func randoParagraph() string {
	par := randomdata.Paragraph()
	if len(par) > 256 {
		par = par[:255]
	}
	return par
}

func fakeMessage(userID string) *messaging.Message {
	return &messaging.Message{
		UserId:      userID,
		Title:       randomdata.Paragraph()[:10],
		Data:        randoParagraph(),
		Link:        fmt.Sprintf("https://%s", randomdata.RandStringRunes(32)),
		Seen:        false,
		Save:        true,
		Type:        randomType(),
		SendMethods: []messaging.SendMethod{randomSendMethod()},
		Details: map[string]string{
			"time": time.Now().String(),
			"from": randomdata.Email(),
		},
	}
}

var _ = Describe("Sending messages @sending", func() {
	var (
		sendReq *messaging.Message
		ctx     context.Context
		userID  = fmt.Sprint(randomdata.Number(100, 999))
	)

	BeforeEach(func() {
		sendReq = fakeMessage(userID)
		ctx = context.Background()
	})

	Describe("Sending message with malformed request", func() {
		It("should fail if request is nil", func() {
			sendReq = nil
			sendRes, err := MessagingAPI.SendMessage(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail if user id is missing", func() {
			sendReq.UserId = ""
			sendRes, err := MessagingAPI.SendMessage(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail if message title is missing", func() {
			sendReq.Title = ""
			sendRes, err := MessagingAPI.SendMessage(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail if message data is missing", func() {
			sendReq.Data = ""
			sendRes, err := MessagingAPI.SendMessage(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail if message details is missing", func() {
			sendReq.Details = nil
			sendRes, err := MessagingAPI.SendMessage(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail if message sendmethod is missing", func() {
			sendReq.SendMethods = nil
			sendRes, err := MessagingAPI.SendMessage(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
		It("should fail if message sendmethod is unknown", func() {
			sendReq.SendMethods = []messaging.SendMethod{messaging.SendMethod_SEND_METHOD_UNSPECIFIED}
			sendRes, err := MessagingAPI.SendMessage(ctx, sendReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(sendRes).Should(BeNil())
		})
	})

	Describe("Sending message with well-formed request", func() {

		Describe("Sending a random message", func() {
			var messageID string

			It("should succeed", func() {
				sendRes, err := MessagingAPI.SendMessage(ctx, sendReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(sendRes).ShouldNot(BeNil())
				_, err = strconv.Atoi(sendRes.MessageId)
				Expect(err).ShouldNot(HaveOccurred())

				messageID = sendRes.MessageId
			})

			Describe("The message should be sent and saved in table", func() {
				It("should be available in table", func() {
					msg := &Message{}
					err := MessagingServer.SQLDBWrites.Table(messages).
						First(msg, "id=? AND user_id=?", messageID, userID).Error
					Expect(err).ShouldNot(HaveOccurred())
				})
			})
		})

		for _, sendMethod := range messaging.SendMethod_value {
			sendMethod := sendMethod
			Describe("Different send methods", func() {

				if messaging.SendMethod(sendMethod) == messaging.SendMethod_SEND_METHOD_UNSPECIFIED {
					It("should fail because send method is unknown", func() {
						sendReq.SendMethods = []messaging.SendMethod{messaging.SendMethod(sendMethod)}
						sendRes, err := MessagingAPI.SendMessage(ctx, sendReq)
						Expect(err).Should(HaveOccurred())
						Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
						Expect(sendRes).Should(BeNil())
					})
					return
				}

				var messageID string

				It("should succeed", func() {
					sendReq.SendMethods = []messaging.SendMethod{messaging.SendMethod(sendMethod)}
					sendRes, err := MessagingAPI.SendMessage(ctx, sendReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(sendRes).ShouldNot(BeNil())
					_, err = strconv.Atoi(sendRes.MessageId)
					Expect(err).ShouldNot(HaveOccurred())

					messageID = sendRes.MessageId
				})

				Describe("The message should be sent and saved in table", func() {
					It("should available in table", func() {
						msg := &Message{}
						err := MessagingServer.SQLDBWrites.Table(messages).
							First(msg, "ID=? AND user_id=?", messageID, userID).Error
						Expect(err).ShouldNot(HaveOccurred())
						msgPB, err := GetMessagePB(msg)
						Expect(err).ShouldNot(HaveOccurred())
						Expect(messaging.SendMethod(sendMethod)).Should(BeElementOf(msgPB.SendMethods))
					})
				})
			})
		}
	})
})
