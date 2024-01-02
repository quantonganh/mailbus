package mailbus

// SubscriptionService is the interface that wraps methods related to subscribe function
type SubscriptionService interface {
	FindByEmail(email string) (*Subscription, error)
	Insert(s *Subscription) error
	Update(email, token string) error
	FindByToken(token string) (*Subscription, error)
	FindByStatus(status string) ([]Subscription, error)
	Subscribe(token string) (string, error)
	Unsubscribe(email string) error
}

// Subscription represents a subscriber
type Subscription struct {
	ID     int    `storm:"id,increment"`
	Email  string `storm:"unique"`
	Token  string `storm:"index"`
	Status string `storm:"index"`
}

// Subscribe status
const (
	StatusPendingConfirmation = "pending_confirmation"
	StatusActive              = "active"
	StatusUnsubscribed        = "unsubscribed"
)

// NewSubscription returns new subscriber
func NewSubscription(email, token, status string) *Subscription {
	return &Subscription{
		Email:  email,
		Token:  token,
		Status: status,
	}
}

type SubscriptionRequest struct {
	URL   string `json:"url"`
	Email string `json:"email"`
}
