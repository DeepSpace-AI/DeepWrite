package mailer

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"strings"
	"time"
)

const defaultSMTPTimeout = 10 * time.Second

var (
	ErrSMTPHostRequired   = errors.New("smtp host is required")
	ErrSMTPPortInvalid    = errors.New("smtp port must be greater than 0")
	ErrSMTPFromRequired   = errors.New("smtp from mail is required")
	ErrRecipientRequired  = errors.New("at least one recipient is required")
	ErrMessageBodyMissing = errors.New("either text_body or html_body is required")
	ErrTLSModeInvalid     = errors.New("invalid smtp tls mode")
	ErrSTARTTLSRequired   = errors.New("smtp server does not support STARTTLS")
)

type SMTPSender struct {
	cfg SMTPConfig
}

const (
	tlsModeAuto     = "auto"
	tlsModeStartTLS = "starttls"
	tlsModeSSL      = "ssl"
	tlsModeNone     = "none"
)

func NewSMTPSender(cfg SMTPConfig) (*SMTPSender, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, ErrSMTPHostRequired
	}
	if cfg.Port <= 0 {
		return nil, ErrSMTPPortInvalid
	}
	if strings.TrimSpace(cfg.FromMail) == "" {
		return nil, ErrSMTPFromRequired
	}

	mode := normalizeTLSMode(cfg.TLSMode)
	if mode == "" {
		return nil, ErrTLSModeInvalid
	}
	cfg.TLSMode = mode

	return &SMTPSender{cfg: cfg}, nil
}

func (s *SMTPSender) Send(ctx context.Context, message Message) error {
	if len(message.Recipients()) == 0 {
		return ErrRecipientRequired
	}
	if strings.TrimSpace(message.TextBody) == "" && strings.TrimSpace(message.HTMLBody) == "" {
		return ErrMessageBodyMissing
	}

	mimeBody, err := buildMIMEMessage(s.cfg, message)
	if err != nil {
		return err
	}

	addr := net.JoinHostPort(s.cfg.Host, fmt.Sprintf("%d", s.cfg.Port))

	dialer := &net.Dialer{Timeout: defaultSMTPTimeout}

	var (
		conn net.Conn
	)

	if s.shouldUseImplicitTLS() {
		// Port 465 uses implicit TLS and does not support plain SMTP handshake first.
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12})
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(defaultSMTPTimeout))
	}

	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return err
	}
	defer c.Quit()

	if s.shouldUseStartTLS() {
		if ok, _ := c.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{ServerName: s.cfg.Host}
			if err := c.StartTLS(tlsConfig); err != nil {
				return err
			}
		} else if s.requiresStartTLS() {
			return ErrSTARTTLSRequired
		}
	}

	if strings.TrimSpace(s.cfg.Username) != "" {
		auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
		if err := c.Auth(auth); err != nil {
			return err
		}
	}

	if err := c.Mail(s.cfg.FromMail); err != nil {
		return err
	}
	for _, rcpt := range message.Recipients() {
		if err := c.Rcpt(strings.TrimSpace(rcpt)); err != nil {
			return err
		}
	}

	wc, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := wc.Write(mimeBody); err != nil {
		_ = wc.Close()
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

func buildMIMEMessage(cfg SMTPConfig, m Message) ([]byte, error) {
	boundary := fmt.Sprintf("dw-mail-%d", time.Now().UnixNano())

	fromHeader := strings.TrimSpace(cfg.FromMail)
	if strings.TrimSpace(cfg.FromName) != "" {
		fromHeader = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("UTF-8", cfg.FromName), strings.TrimSpace(cfg.FromMail))
	}

	headers := map[string]string{
		"From":         fromHeader,
		"To":           strings.Join(trimmedList(m.To), ", "),
		"Subject":      mime.QEncoding.Encode("UTF-8", strings.TrimSpace(m.Subject)),
		"MIME-Version": "1.0",
	}
	if len(trimmedList(m.Cc)) > 0 {
		headers["Cc"] = strings.Join(trimmedList(m.Cc), ", ")
	}

	var buf bytes.Buffer
	for k, v := range headers {
		if strings.TrimSpace(v) == "" {
			continue
		}
		buf.WriteString(k + ": " + v + "\r\n")
	}

	hasText := strings.TrimSpace(m.TextBody) != ""
	hasHTML := strings.TrimSpace(m.HTMLBody) != ""

	if hasText && hasHTML {
		buf.WriteString("Content-Type: multipart/alternative; boundary=" + boundary + "\r\n\r\n")
		appendPart(&buf, boundary, "text/plain; charset=UTF-8", m.TextBody)
		appendPart(&buf, boundary, "text/html; charset=UTF-8", m.HTMLBody)
		buf.WriteString("--" + boundary + "--\r\n")
		return buf.Bytes(), nil
	}

	if hasHTML {
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		qp := quotedprintable.NewWriter(&buf)
		if _, err := qp.Write([]byte(m.HTMLBody)); err != nil {
			return nil, err
		}
		if err := qp.Close(); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	qp := quotedprintable.NewWriter(&buf)
	if _, err := qp.Write([]byte(m.TextBody)); err != nil {
		return nil, err
	}
	if err := qp.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func appendPart(buf *bytes.Buffer, boundary, contentType, body string) {
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString("Content-Type: " + contentType + "\r\n")
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	qp := quotedprintable.NewWriter(buf)
	_, _ = qp.Write([]byte(body))
	_ = qp.Close()
	buf.WriteString("\r\n")
}

func trimmedList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, item := range values {
		v := strings.TrimSpace(item)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func normalizeTLSMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", tlsModeAuto:
		return tlsModeAuto
	case tlsModeStartTLS:
		return tlsModeStartTLS
	case tlsModeSSL:
		return tlsModeSSL
	case tlsModeNone:
		return tlsModeNone
	default:
		return ""
	}
}

func (s *SMTPSender) shouldUseImplicitTLS() bool {
	if s.cfg.TLSMode == tlsModeSSL {
		return true
	}
	return s.cfg.TLSMode == tlsModeAuto && s.cfg.Port == 465
}

func (s *SMTPSender) shouldUseStartTLS() bool {
	if s.cfg.TLSMode == tlsModeStartTLS {
		return true
	}
	return s.cfg.TLSMode == tlsModeAuto && s.cfg.Port != 465
}

func (s *SMTPSender) requiresStartTLS() bool {
	return s.cfg.TLSMode == tlsModeStartTLS
}
