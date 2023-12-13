package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

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
	err := ss.db.sqlDB.QueryRow("SELECT * FROM subscriptions WHERE email = ?", email).
		Scan(&s.Email, &s.Token, &s.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Subscription not found
		}
		return nil, fmt.Errorf("failed to find by email %s: %w", email, err)
	}
	return &s, nil
}

// Insert inserts new subscription into stormDB
func (ss *subscriptionService) Insert(s *mailbus.Subscription) error {
	_, err := ss.db.sqlDB.Exec("INSERT INTO subscriptions (email, token, status) VALUES (?, ?, ?)",
		s.Email, s.Token, s.Status)
	if err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}
	return nil
}

// Update updates subscription status and new token
func (ss *subscriptionService) Update(email, token string) error {
	_, err := ss.db.sqlDB.Exec("UPDATE subscriptions SET status = ?, token = ? WHERE email = ?",
		mailbus.StatusPending, token, email)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}
	return nil
}

// FindByToken finds subscription by token
func (ss *subscriptionService) FindByToken(token string) (*mailbus.Subscription, error) {
	var s mailbus.Subscription
	err := ss.db.sqlDB.QueryRow("SELECT * FROM subscriptions WHERE token = ?", token).
		Scan(&s.Email, &s.Token, &s.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Subscription not found
		}
		return nil, fmt.Errorf("failed to find by token: %w", err)
	}
	return &s, nil
}

// FindByStatus finds subscription by status
func (ss *subscriptionService) FindByStatus(status string) ([]mailbus.Subscription, error) {
	var subscriptions []mailbus.Subscription
	rows, err := ss.db.sqlDB.Query("SELECT * FROM subscriptions WHERE status = ?", status)
	if err != nil {
		return nil, fmt.Errorf("failed to find by status: %w", err)
	}
	defer rows.Close()

	// Iterate over the rows and populate the subscriptions slice
	for rows.Next() {
		var s mailbus.Subscription
		err := rows.Scan(&s.Email, &s.Token, &s.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		subscriptions = append(subscriptions, s)
	}

	return subscriptions, nil
}

// Subscribe subscribes to newsletter
func (ss *subscriptionService) Subscribe(token string) error {
	_, err := ss.db.sqlDB.Exec("UPDATE subscriptions SET status = ? WHERE token = ?",
		mailbus.StatusSubscribed, token)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

// Unsubscribe unsubscribes from newsletter
func (ss *subscriptionService) Unsubscribe(email string) error {
	_, err := ss.db.sqlDB.Exec("UPDATE subscriptions SET status = ? WHERE email = ?",
		mailbus.StatusUnsubscribed, email)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}
