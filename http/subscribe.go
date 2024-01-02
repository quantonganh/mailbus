package http

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/hlog"

	"github.com/quantonganh/mailbus"
)

func (s *Server) subscriptionsHandler(w http.ResponseWriter, r *http.Request) error {
	var req *mailbus.SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	email := req.Email

	token := s.NewsletterService.GenerateNewUUID()
	newSubscription := mailbus.NewSubscription(email, token, mailbus.StatusPendingConfirmation)

	logger := hlog.FromRequest(r)
	subscribe, err := s.SubscriptionService.FindByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info().Msg("Sending confirmation email")
			if err := s.NewsletterService.SendConfirmationEmail(email, req.URL, token); err != nil {
				return err
			}

			logger.Info().Msgf("Saving new subscriber %+v into the database", newSubscription)
			if err := s.SubscriptionService.Insert(newSubscription); err != nil {
				return err
			}

			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		logger.Info().Msgf("Found subscriber %+v in the database", subscribe)
		switch subscribe.Status {
		case mailbus.StatusPendingConfirmation:
			w.WriteHeader(http.StatusUnauthorized)
		case mailbus.StatusActive:
			w.WriteHeader(http.StatusConflict)
		default:
			if err := s.NewsletterService.SendConfirmationEmail(email, req.URL, token); err != nil {
				return err
			}

			logger.Info().Msgf("Updating status to %s", mailbus.StatusPendingConfirmation)
			if err := s.SubscriptionService.Update(email, token); err != nil {
				return err
			}

			w.WriteHeader(http.StatusOK)
		}
	}

	return nil
}

func (s *Server) confirmHandler(w http.ResponseWriter, r *http.Request) error {
	token := r.URL.Query().Get("token")
	if len(token) == 0 {
		return errors.New("token is not present")
	}

	email, err := s.SubscriptionService.Subscribe(token)
	if err != nil {
		return err
	}

	if err := s.NewsletterService.SendThankYouEmail(email); err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)

	return nil
}
