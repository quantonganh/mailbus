package bolt

import (
	"github.com/go-errors/errors"
	"github.com/quantonganh/mailbus"
)

type subscriptionService struct {
	db *DB
}

func NewSubscriptionService(db *DB) mailbus.SubscriptionService {
	return &subscriptionService{
		db: db,
	}
}

// FindByEmail finds a subscription by email
func (ss *subscriptionService) FindByEmail(email string) (*mailbus.Subscription, error) {
	var s mailbus.Subscription
	if err := ss.db.stormDB.One("Email", email, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

// Insert inserts new subscription into stormDB
func (ss *subscriptionService) Insert(s *mailbus.Subscription) error {
	if err := ss.db.stormDB.Save(s); err != nil {
		return errors.Errorf("failed to save: %v", err)
	}

	return nil
}

// Update updates subscription status and new token
func (ss *subscriptionService) Update(email, token string) error {
	s, err := ss.FindByEmail(email)
	if err != nil {
		return err
	}

	s.Status = mailbus.StatusPending
	s.Token = token
	if err := ss.db.stormDB.Save(s); err != nil {
		return errors.Errorf("failed to save: %v", err)
	}

	return nil
}

// FindByToken finds subscription by token
func (ss *subscriptionService) FindByToken(token string) (*mailbus.Subscription, error) {
	var s mailbus.Subscription
	if err := ss.db.stormDB.One("Token", token, &s); err != nil {
		return nil, errors.Errorf("failed to find by token: %v", err)
	}

	return &s, nil
}

// FindByStatus finds subscription by status
func (ss *subscriptionService) FindByStatus(status string) ([]mailbus.Subscription, error) {
	var subscribes []mailbus.Subscription
	if err := ss.db.stormDB.Find("Status", status, &subscribes); err != nil {
		return nil, errors.Errorf("failed to find by status: %v", err)
	}

	return subscribes, nil
}

// Subscribe subscribes to newsletter
func (ss *subscriptionService) Subscribe(token string) error {
	s, err := ss.FindByToken(token)
	if err != nil {
		return err
	}

	s.Status = mailbus.StatusSubscribed
	if err := ss.db.stormDB.Save(s); err != nil {
		return err
	}

	return nil
}

// Unsubscribe unsubscribes from newsletter
func (ss *subscriptionService) Unsubscribe(email string) error {
	s, err := ss.FindByEmail(email)
	if err != nil {
		return errors.Errorf("failed to find by email: %v", err)
	}

	s.Status = mailbus.StatusUnsubscribed
	if err := ss.db.stormDB.Save(s); err != nil {
		return errors.Errorf("failed to save: %v", err)
	}

	return nil
}
