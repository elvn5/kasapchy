package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type SMTP struct {
	Address, Host, Username, Password, From string
	Local                                   bool
}

func (s SMTP) SendCode(ctx context.Context, email, code string) error {
	if strings.ContainsAny(s.From+email, "\r\n") {
		return fmt.Errorf("invalid mail address")
	}
	conn, e := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", s.Address)
	if e != nil {
		return e
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))
	client, e := smtp.NewClient(conn, s.Host)
	if e != nil {
		return e
	}
	defer client.Close()
	if !s.Local {
		if e = client.StartTLS(&tls.Config{ServerName: s.Host, MinVersion: tls.VersionTLS12}); e != nil {
			return e
		}
	}
	if s.Username != "" {
		if e = client.Auth(smtp.PlainAuth("", s.Username, s.Password, s.Host)); e != nil {
			return e
		}
	}
	if e = client.Mail(s.From); e != nil {
		return e
	}
	if e = client.Rcpt(email); e != nil {
		return e
	}
	w, e := client.Data()
	if e != nil {
		return e
	}
	_, e = fmt.Fprintf(w, "From: %s\r\nTo: %s\r\nSubject: Kasapchy verification code\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nКод входа в Касапчы: %s\r\nКод действует 10 минут. Не сообщайте его другим.\r\n", s.From, email, code)
	if e != nil {
		return e
	}
	if e = w.Close(); e != nil {
		return e
	}
	return client.Quit()
}
