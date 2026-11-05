package cloudy

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/mailer"
)

// smtpSender sends mail with net/smtp and also satisfies mailer.Mailer for tenant apps.
type smtpSender struct {
	cfg SMTPConfig
}

func newSMTPSender(cfg SMTPConfig) *smtpSender {
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	if cfg.From == "" {
		cfg.From = cfg.Username
	}
	return &smtpSender{cfg: cfg}
}

func (s *smtpSender) Send(to string, subject string, body mailer.MessageBody) error {
	html, err := body.AsHTML()
	if err != nil {
		return err
	}
	text, err := body.AsText()
	if err != nil {
		return err
	}
	return s.send(to, subject, text, html)
}

func (s *smtpSender) send(to, subject, text, html string) error {
	if s.cfg.Host == "" {
		log.Printf("[mail/stdio] to=%s subject=%s body=%s", to, subject, text)
		return nil
	}

	fromHeader := s.cfg.From
	if s.cfg.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.From)
	}

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", fromHeader))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	if html != "" {
		boundary := "cloudy-boundary"
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary))
		if text != "" {
			msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
			msg.WriteString(text)
			msg.WriteString("\r\n")
		}
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		msg.WriteString(html)
		msg.WriteString("\r\n")
		msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		msg.WriteString(text)
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	payload := []byte(msg.String())

	if s.cfg.Port == 465 {
		return sendSMTPTLS(s.cfg.Host, addr, auth, s.cfg.From, to, payload)
	}
	return smtp.SendMail(addr, auth, s.cfg.From, []string{to}, payload)
}

func sendSMTPTLS(serverName, addr string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: serverName})
	if err != nil {
		return err
	}
	defer conn.Close()

	host, _, _ := net.SplitHostPort(addr)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func (a *CloudyApp) sendMail(to, subject, text, html string) error {
	return a.mailer.send(to, subject, text, html)
}
