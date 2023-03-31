package utility

import (
	"io"

	"github.com/go-mail/mail"
)

type EmbedFile struct {
	Name   string
	Reader io.Reader
}

func SendEmail(receiver, subject string, writerFunc func(w io.Writer) error, embedFiles ...*EmbedFile) error {
	m := mail.NewMessage()
	m.SetHeader("From", Config.Email.Account)
	m.SetHeader("To", receiver)
	m.SetHeader("Subject", subject)
	m.SetBodyWriter("text/html", writerFunc)
	for _, embedFile := range embedFiles {
		m.EmbedReader(embedFile.Name, embedFile.Reader)
	}
	d := mail.NewDialer("smtp.gmail.com", 465, Config.Email.Account, Config.Email.Password)
	return d.DialAndSend(m)
}
