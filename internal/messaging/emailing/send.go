package emailing

import (
	"github.com/gidyon/services/pkg/api/messaging/emailing"
	"gopkg.in/gomail.v2"
)

func (api *emailingAPIServer) sendEmail(email *emailing.Email) error {
	m := gomail.NewMessage()
	m.SetHeader("From", email.From)
	m.SetHeader("To", email.Destinations...)
	m.SetHeader("Subject", email.Subject)
	m.SetBody(email.BodyContentType, email.Body)

	return api.dialer.DialAndSend(m)
}
