package mailbus

// NewsletterService is the interface that wraps methods related to SMTP
type NewsletterService interface {
	SendConfirmationEmail(to, url, token string) error
	SendThankYouEmail(to string) error
	SendNewsletter(subscribers []Subscriber, subject, body string)
	GenerateNewUUID() string
	GetHMACSecret() string
}

type EmailNewsletterRequest struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
