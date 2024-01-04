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
	err := ss.db.sqlDB.QueryRow("SELECT email, status FROM subscriptions WHERE email = ?", email).
		Scan(&s.Email, &s.Status)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Insert inserts new subscription into stormDB
func (ss *subscriptionService) Insert(s *mailbus.Subscription) error {
	tx, err := ss.db.sqlDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start a transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()

	result, err := tx.Exec("INSERT INTO subscriptions (email, status) VALUES (?, ?)",
		s.Email, s.Status)
	if err != nil {
		return fmt.Errorf("failed to insert into subscriptions table: %w", err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	_, err = tx.Exec("INSERT INTO subscription_tokens (subscription_token, subscriber_id) VALUES (?, ?)",
		s.Token, lastInsertID)
	if err != nil {
		return fmt.Errorf("failed to insert into subscriptions table: %w", err)
	}

	return nil
}

// Update updates subscription status and new token
func (ss *subscriptionService) Update(email, token string) error {
	_, err := ss.db.sqlDB.Exec("UPDATE subscriptions SET status = ?, token = ? WHERE email = ?",
		mailbus.StatusPendingConfirmation, token, email)
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
func (ss *subscriptionService) Confirm(token string) (string, error) {
	const op = "subscriptionService.Confirm"

	row := ss.db.sqlDB.QueryRow(`
		SELECT s.id, s.email
		FROM subscription_tokens t
		JOIN subscriptions s ON t.subscriber_id = s.id
		WHERE t.subscription_token = ?`, token)
	var (
		subscriberID int64
		email        string
	)
	if err := row.Scan(&subscriberID, &email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", &mailbus.Error{
				Code: mailbus.ErrNotFound,
			}
		} else {
			return "", &mailbus.Error{
				Code: mailbus.ErrInternal,
				Op:   op,
				Err:  err,
			}
		}
	}

	_, err := ss.db.sqlDB.Exec("UPDATE subscriptions SET status = ? WHERE id = ?", mailbus.StatusActive, subscriberID)
	if err != nil {
		return "", fmt.Errorf("failed to update status: %w", err)
	}
	return email, nil
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
