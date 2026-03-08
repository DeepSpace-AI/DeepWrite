package mailer

import "context"

type Sender interface {
	Send(ctx context.Context, message Message) error
}

type SMTPConfig struct {
	Host     string
	Port     int
	TLSMode  string
	Username string
	Password string
	FromName string
	FromMail string
}

type Message struct {
	To       []string
	Cc       []string
	Bcc      []string
	Subject  string
	TextBody string
	HTMLBody string
}

func (m Message) Recipients() []string {
	all := make([]string, 0, len(m.To)+len(m.Cc)+len(m.Bcc))
	all = append(all, m.To...)
	all = append(all, m.Cc...)
	all = append(all, m.Bcc...)
	return all
}
