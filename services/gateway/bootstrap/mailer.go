package bootstrap

import (
	"strings"

	"github.com/deepwrite/serivces/gateway/pkg/config"
	"github.com/deepwrite/serivces/gateway/pkg/mailer"
)

var Mailer mailer.Sender

func SetupMailer(cfg config.Config) error {
	if !cfg.Mail.Enabled {
		return nil
	}
	if !strings.EqualFold(strings.TrimSpace(cfg.Mail.Provider), "smtp") {
		return nil
	}

	sender, err := mailer.NewSMTPSender(mailer.SMTPConfig{
		Host:     cfg.Mail.Host,
		Port:     cfg.Mail.Port,
		TLSMode:  cfg.Mail.TLSMode,
		Username: cfg.Mail.Username,
		Password: cfg.Mail.Password,
		FromName: cfg.Mail.FromName,
		FromMail: cfg.Mail.FromMail,
	})
	if err != nil {
		return err
	}

	Mailer = sender
	mailer.SetDefault(sender)
	return nil
}

func GetMailer() mailer.Sender {
	return Mailer
}
