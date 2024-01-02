package gmail

import (
	"fmt"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/matcornic/hermes/v2"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/gomail.v2"

	"github.com/quantonganh/mailbus"
)

type newsletterService struct {
	ServerURL string
	*mailbus.Config
	mailbus.SubscriptionService
	*cron.Cron
}

// NewNewsletterService returns new newsletter service
func NewNewsletterService(config *mailbus.Config, serverURL string, subscriptionService mailbus.SubscriptionService) mailbus.NewsletterService {
	return &newsletterService{
		Config:              config,
		ServerURL:           serverURL,
		SubscriptionService: subscriptionService,
		Cron:                cron.New(cron.WithLogger(cron.DefaultLogger)),
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
func (ns *newsletterService) SendNewsletter(content string) {
	_, err := ns.Cron.AddFunc(ns.Config.Newsletter.Cron.Spec, func() {

		subscribers, err := ns.SubscriptionService.FindByStatus(mailbus.StatusActive)
		if err != nil {
			sentry.CaptureException(err)
		}

		for _, s := range subscribers {
			if err := ns.sendEmail(s.Email, fmt.Sprintf("%s newsletter", ns.Config.Newsletter.Product.Name), content); err != nil {
				sentry.CaptureException(err)
			}
		}
	})
	if err != nil {
		sentry.CaptureException(err)
	}

	ns.Cron.Start()
}

// Stop stops newsletter service
func (ns *newsletterService) Stop() error {
	ctx := ns.Cron.Stop()
	log.Println("Shutting down cron...")
	select {
	case <-time.After(10 * time.Second):
		return errors.New("cron forced to shutdown")
	case <-ctx.Done():
		log.Println("Cron exiting...")
		return ctx.Err()
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
