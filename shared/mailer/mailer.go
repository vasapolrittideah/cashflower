package mailer

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog"
	"gopkg.in/gomail.v2"
)

// Mailer represents an email sender.
type Mailer struct {
	config *mailerConfig
	dialer *gomail.Dialer
}

// Email represents an email message.
type Email struct {
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []string
	Embeds      []string
}

// NewMailer creates a new Mailer instance with the given configuration.
func NewMailer(logger *zerolog.Logger) *Mailer {
	cfg := newMailerConfig(logger)

	if err := cfg.validate(); err != nil {
		logger.Fatal().Err(err).Msg("failed to validate Mailer configuration")
	}

	dialer := gomail.NewDialer(
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
	)

	return &Mailer{
		config: cfg,
		dialer: dialer,
	}
}

// Send sends a single email.
func (m *Mailer) Send(email Email) error {
	if len(email.To) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	msg := gomail.NewMessage()
	m.setEmailMessage(msg, email)

	return m.dialer.DialAndSend(msg)
}

// SendBulk sends multiple emails in a single operation.
func (m *Mailer) SendBulk(emails []Email) error {
	sender, err := m.dialer.Dial()
	if err != nil {
		return err
	}
	defer sender.Close()

	for i, email := range emails {
		msg := gomail.NewMessage()
		m.setEmailMessage(msg, email)

		if err := gomail.Send(sender, msg); err != nil {
			return fmt.Errorf("failed to send email %d: %w", i+1, err)
		}

		msg.Reset()
	}

	return nil
}

// SendSimple sends a simple text email.
func (m *Mailer) SendSimple(to []string, subject, body string) error {
	return m.Send(Email{
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

// SendHTML sends an HTML email.
func (m *Mailer) SendHTML(to []string, subject, htmlBody string) error {
	return m.Send(Email{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
	})
}

// SendWithAttachment sends an email with attachments.
func (m *Mailer) SendWithAttachment(to []string, subject, body string, attachments []string) error {
	return m.Send(Email{
		To:          to,
		Subject:     subject,
		Body:        body,
		Attachments: attachments,
	})
}

func (m *Mailer) setEmailMessage(msg *gomail.Message, email Email) {
	// Set headers
	msg.SetHeader("From", m.config.From)
	msg.SetHeader("To", email.To...)

	if len(email.Cc) > 0 {
		msg.SetHeader("Cc", email.Cc...)
	}

	if len(email.Bcc) > 0 {
		msg.SetHeader("Bcc", email.Bcc...)
	}

	msg.SetHeader("Subject", email.Subject)

	// Set body
	if email.HTMLBody != "" {
		msg.SetBody("text/html", email.HTMLBody)
		if email.Body != "" {
			msg.AddAlternative("text/plain", email.Body)
		}
	} else {
		msg.SetBody("text/plain", email.Body)
	}

	// Add attachments
	for _, attachment := range email.Attachments {
		msg.Attach(attachment)
	}

	// Add embedded images
	for _, embed := range email.Embeds {
		msg.Embed(embed)
	}
}

// mailerConfig holds SMTP configuration for sending emails.
type mailerConfig struct {
	Host     string `env:"SMTP_HOST"`
	Port     int    `env:"SMTP_PORT"`
	Username string `env:"SMTP_USERNAME"`
	Password string `env:"SMTP_PASSWORD"`
	From     string `env:"SMTP_FROM"`
}

// newMailerConfig creates a MailerConfig instance from environment variables.
func newMailerConfig(logger *zerolog.Logger) *mailerConfig {
	cfg, err := env.ParseAs[mailerConfig]()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse environment variables")
	}

	return &cfg
}

// validate checks if the Mailer configuration is valid.
func (c *mailerConfig) validate() error {
	if c.Host == "" {
		return fmt.Errorf("missing SMTP_HOST environment variable")
	}
	if c.Port == 0 {
		return fmt.Errorf("missing SMTP_PORT environment variable")
	}
	if c.Username == "" {
		return fmt.Errorf("missing SMTP_USERNAME environment variable")
	}
	if c.Password == "" {
		return fmt.Errorf("missing SMTP_PASSWORD environment variable")
	}
	if c.From == "" {
		return fmt.Errorf("missing SMTP_FROM environment variable")
	}

	return nil
}
