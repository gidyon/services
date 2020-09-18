package emailing

import (
	"io/ioutil"

	"github.com/gidyon/services/pkg/api/messaging/emailing"
	"github.com/gidyon/services/pkg/utils/errs"
	"gopkg.in/gomail.v2"
)

func (api *emailingAPIServer) sendEmail(email *emailing.Email) error {
	m := gomail.NewMessage()
	m.SetHeader("From", email.From)
	m.SetHeader("To", email.Destinations...)
	m.SetHeader("Subject", email.Subject)
	m.SetBody(email.BodyContentType, email.Body)

	var err error

	// Create files
	for _, attachment := range email.Attachments {
		err = ioutil.WriteFile(attachment.Filename, attachment.Data, 0666)
		if err != nil {
			return errs.WriteFailed(err)
		}
	}

	// Send attachements
	for _, attachment := range email.Attachments {
		m.Attach(attachment.Filename, gomail.Rename(attachment.FilenameOverride))
	}

	return api.dialer.DialAndSend(m)
}
