package mailbus

// SubscriptionService is the interface that wraps methods related to subscribe function
type SubscriptionService interface {
	FindByEmail(email string) (*Subscription, error)
	Insert(s *Subscription) error
	Update(email, token string) error
	FindByToken(token string) (*Subscription, error)
	FindByStatus(status string) ([]Subscription, error)
	Subscribe(token string) error
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
	StatusPending      = "pending"
	StatusSubscribed   = "subscribed"
	StatusUnsubscribed = "unsubscribed"
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
	Email string `json:"email"`
	Token string `json:"token"`
}

type SubscriptionResponse struct {
	Message string `json:"message"`
}
