package platform

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type EmailSender interface {
	Send(ctx context.Context, to string, subject string, bodyText string) error
}

type SMTPSender struct {
	Host       string
	Port       int
	Username   string
	Password   string
	FromEmail  string
	FromName   string
	StartTLS   bool
	Timeout    time.Duration
	ServerName string
}

func (s *SMTPSender) Send(ctx context.Context, to string, subject string, bodyText string) error {
	if strings.TrimSpace(s.Host) == "" {
		return fmt.Errorf("smtp host is required")
	}
	if s.Port <= 0 {
		return fmt.Errorf("smtp port is required")
	}
	if strings.TrimSpace(s.FromEmail) == "" {
		return fmt.Errorf("smtp from email is required")
	}
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("email recipient is required")
	}

	timeout := s.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	dialer := &net.Dialer{Timeout: timeout}
	addr := net.JoinHostPort(s.Host, fmt.Sprintf("%d", s.Port))
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return err
	}
	defer c.Quit()

	if s.StartTLS {
		serverName := strings.TrimSpace(s.ServerName)
		if serverName == "" {
			serverName = s.Host
		}
		cfg := &tls.Config{ServerName: serverName, MinVersion: tls.VersionTLS12}
		if err := c.StartTLS(cfg); err != nil {
			return err
		}
	}

	if strings.TrimSpace(s.Username) != "" {
		auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)
		if err := c.Auth(auth); err != nil {
			return err
		}
	}

	from := s.FromEmail
	if err := c.Mail(from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	fromHeader := from
	if strings.TrimSpace(s.FromName) != "" {
		fromHeader = fmt.Sprintf("%s <%s>", s.FromName, from)
	}

	msg := strings.Builder{}
	msg.WriteString("From: " + fromHeader + "\r\n")
	msg.WriteString("To: " + to + "\r\n")
	msg.WriteString("Subject: " + strings.ReplaceAll(subject, "\n", " ") + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	msg.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(bodyText)
	if !strings.HasSuffix(bodyText, "\n") {
		msg.WriteString("\n")
	}

	_, err = w.Write([]byte(msg.String()))
	return err
}
