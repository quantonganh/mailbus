package mailbus

import "time"

// Subscribe status
const (
	StatusPendingConfirmation = "pending_confirmation"
	StatusActive              = "active"
	StatusUnsubscribed        = "unsubscribed"
)

// SubscriptionService is the interface that wraps methods related to subscribe function
type SubscriptionService interface {
	FindByEmail(email string) (*Subscriber, error)
	Insert(s *Subscription) error
	Update(email, token string) error
	FindByStatus(status string) ([]Subscriber, error)
	Confirm(token string) (string, error)
	Unsubscribe(email string) error
}

// Subscriber represents a subscriber
type Subscriber struct {
	ID           int    `storm:"id,increment"`
	Email        string `storm:"unique"`
	Status       string `storm:"index"`
	SubscribedAt time.Time
}

type Subscription struct {
	Email  string
	Status string
	Token  string
}

// NewSubscription returns new subscriber
func NewSubscription(email, status, token string) *Subscription {
	return &Subscription{
		Email:  email,
		Status: status,
		Token:  token,
	}
}

type SubscriptionRequest struct {
	URL   string `json:"url"`
	Email string `json:"email"`
}
