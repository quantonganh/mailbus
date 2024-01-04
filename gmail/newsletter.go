package gmail

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/matcornic/hermes/v2"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/gomail.v2"

	"github.com/quantonganh/mailbus"
)

type newsletterService struct {
	ServerURL string
	*mailbus.Config
}

// NewNewsletterService returns new newsletter service
func NewNewsletterService(config *mailbus.Config, serverURL string) mailbus.NewsletterService {
	return &newsletterService{
		Config:    config,
		ServerURL: serverURL,
	}
}

// SendConfirmationEmail sends a confirmation email
func (ns *newsletterService) SendConfirmationEmail(to, url, token string) error {
	h := hermes.Hermes{
		Product: hermes.Product{
			Name: ns.Config.Newsletter.Product.Name,
			Link: ns.ServerURL,
		},
	}

	email := hermes.Email{
		Body: hermes.Body{
			Name: "",
			Intros: []string{
				fmt.Sprintf("Welcome to %s", ns.Config.Newsletter.Product.Name),
			},
			Actions: []hermes.Action{
				{
					Instructions: "",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "Confirm your subscription",
						Link:  fmt.Sprintf("%s/subscribe/confirm?token=%s", url, token),
					},
				},
			},
		},
	}

	emailBody, err := h.GenerateHTML(email)
	if err != nil {
		return errors.Errorf("failed to generate HTML email: %v", err)
	}

	return ns.sendEmail(to, "Confirm subscription", emailBody)
}

// SendThankYouEmail sends a "thank you" email
func (ns *newsletterService) SendThankYouEmail(to string) error {
	h := hermes.Hermes{
		Product: hermes.Product{
			Name: ns.Config.Newsletter.Product.Name,
			Link: ns.ServerURL,
		},
	}

	email := hermes.Email{
		Body: hermes.Body{
			Name: "",
			Intros: []string{
				fmt.Sprintf("Thank you for subscribing to %s", ns.Config.Newsletter.Product.Name),
			},
			Actions: []hermes.Action{
				{
					Instructions: "You will receive updates to your inbox.",
				},
			},
		},
	}

	emailBody, err := h.GenerateHTML(email)
	if err != nil {
		return errors.Errorf("failed to generate HTML email: %v", err)
	}

	return ns.sendEmail(to, "Thank you for subscribing", emailBody)
}

// SendNewsletter sends newsletter
func (ns *newsletterService) SendNewsletter(subscribers []mailbus.Subscriber, subject, body string) {
	for _, s := range subscribers {
		if err := ns.sendEmail(s.Email, subject, body); err != nil {
			sentry.CaptureException(err)
		}
	}
}

func (ns *newsletterService) sendEmail(to string, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", ns.Config.Newsletter.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	d := gomail.NewDialer(ns.Config.SMTP.Host, ns.Config.SMTP.Port, ns.Config.SMTP.Username, ns.Config.SMTP.Password)
	if err := d.DialAndSend(m); err != nil {
		return errors.Errorf("failed to send mail to %s: %v", fmt.Sprintf("%+v\n", to), err)
	}

	return nil
}

func (ns *newsletterService) GenerateNewUUID() string {
	return uuid.NewV4().String()
}

// GetHMACSecret gets HMAC secret from config
func (ns *newsletterService) GetHMACSecret() string {
	return ns.Config.Newsletter.HMAC.Secret
}
