package emailing

import (
	"fmt"
	"io/ioutil"

	"github.com/gidyon/services/pkg/api/messaging/emailing"
	"gopkg.in/gomail.v2"
)

func (api *emailingAPIServer) sendEmail(sendReq *emailing.SendEmailRequest) {
	email := sendReq.GetEmail()

	m := gomail.NewMessage()
	if email.GetDisplayName() != "" {
		m.SetHeader("From", fmt.Sprintf("%s <%s>", email.DisplayName, email.From))
	} else {
		m.SetHeader("From", email.From)
	}
	m.SetHeader("To", email.Destinations...)
	m.SetHeader("Subject", email.Subject)
	m.SetBody(email.BodyContentType, email.Body)

	var err error

	// Create files
	for _, attachment := range email.Attachments {
		err = ioutil.WriteFile(attachment.Filename, attachment.Data, 0666)
		if err != nil {
			api.Logger.Errorf("failed to write attachment: %v", err)
		}
	}

	// Send attachements
	for _, attachment := range email.Attachments {
		m.Attach(attachment.Filename, gomail.Rename(attachment.FilenameOverride))
	}

	err = api.dialer.DialAndSend(m)
	if err != nil {
		api.Logger.Errorf("failed to send email: %v", err)
	}
}
