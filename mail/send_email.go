package mail

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/lh-khanhduy/banco_de_rata/utils"
	"github.com/rs/zerolog/log"
	gomail "github.com/wneessen/go-mail"
)

const (
	smtpAuthAddress = "smtp.gmail.com"
	// smtpServerAddress = "smtp.gmail.com:587"
	smtpPort = 587
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type Mailer struct {
	client *gomail.Client
	from   string
}

// NewMailer creates a reusable mailer
func NewMailer(cfg SMTPConfig) (*Mailer, error) {
	client, err := gomail.NewClient(
		cfg.Host,
		gomail.WithPort(cfg.Port),
		gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
		gomail.WithUsername(cfg.Username),
		gomail.WithPassword(cfg.Password),
		gomail.WithTLSConfig(&tls.Config{ServerName: cfg.Host}),
		gomail.WithTLSPortPolicy(gomail.TLSMandatory),
		gomail.WithTimeout(30*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail client: %w", err)
	}

	return &Mailer{
		client: client,
		from:   cfg.From,
	}, nil
}

// NewMailerFromConfig will return a mailer with from config variable
func NewMailerFromConfig() (*Mailer, error) {
	cfg, err := utils.LoadConfig("..")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load configuration")
		return nil, err
	}
	return NewMailer(SMTPConfig{
		Host:     smtpAuthAddress,
		Port:     smtpPort,
		Username: cfg.EmailSenderName,
		Password: cfg.EmailSenderPassword,
		From:     cfg.EmailSenderAddress,
	})
}

func (m *Mailer) SendEmail(to []string, subject, plainText, htmlBody string, attachFiles []string) error {
	msg := gomail.NewMsg()

	if err := msg.From(m.from); err != nil {
		return fmt.Errorf("invalid From: %w", err)
	}
	if err := msg.To(to...); err != nil {
		return fmt.Errorf("invalid To: %w", err)
	}

	msg.Subject(subject)
	msg.SetBodyString(gomail.TypeTextPlain, plainText)
	if htmlBody != "" {
		msg.AddAlternativeString(gomail.TypeTextHTML, htmlBody)
	}

	for _, file := range attachFiles {
		msg.AttachFile(file)
	}

	if err := m.client.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
