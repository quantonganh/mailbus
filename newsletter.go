package mailbus

// NewsletterService is the interface that wraps methods related to SMTP
type NewsletterService interface {
	SendConfirmationEmail(to, token string) error
	SendThankYouEmail(to string) error
	SendNewsletter(content string)
	GenerateNewUUID() string
	GetHMACSecret() string
	Stop() error
}
