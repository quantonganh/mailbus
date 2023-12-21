package http

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/hlog"

	"github.com/quantonganh/mailbus"
)

const (
	confirmationMessage      = "A confirmation email has been sent to %s. Click the link in the email to confirm and activate your subscription. Check your spam folder if you don't see it within a couple of minutes."
	thankyouMessage          = "Thank you for subscribing to this blog."
	pendingMessage           = "Your subscription status is pending. Please click the confirmation link in your email."
	alreadySubscribedMessage = "You had been subscribed to this blog already."
)

func (s *Server) subscriptionsHandler(w http.ResponseWriter, r *http.Request) error {
	var (
		req  *mailbus.SubscriptionRequest
		resp = new(mailbus.SubscriptionResponse)
	)
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
			if err := s.NewsletterService.SendConfirmationEmail(email, token); err != nil {
				return err
			}

			logger.Info().Msgf("Saving new subscriber %+v into the database", newSubscription)
			if err := s.SubscriptionService.Insert(newSubscription); err != nil {
				return err
			}

			logger.Info().Msg("Rendering the response message")
			resp.Message = fmt.Sprintf(confirmationMessage, newSubscription.Email)
			writeJSONResponse(w, http.StatusOK, resp)
		} else {
			return NewError(err, http.StatusNotFound, fmt.Sprintf("Cannot found email: %s", email))
		}
	} else {
		logger.Info().Msgf("Found subscriber %+v in the database", subscribe)
		switch subscribe.Status {
		case mailbus.StatusPendingConfirmation:
			resp.Message = pendingMessage
			writeJSONResponse(w, http.StatusOK, resp)
		case mailbus.StatusActive:
			resp.Message = alreadySubscribedMessage
			writeJSONResponse(w, http.StatusBadRequest, resp)
		default:
			if err := s.NewsletterService.SendConfirmationEmail(email, token); err != nil {
				return err
			}

			logger.Info().Msgf("Updating status to %s", mailbus.StatusPendingConfirmation)
			if err := s.SubscriptionService.Update(email, token); err != nil {
				return err
			}

			logger.Info().Msg("Rendering the response message")
			resp.Message = fmt.Sprintf(confirmationMessage, email)
			writeJSONResponse(w, http.StatusOK, resp)
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

	writeJSONResponse(w, http.StatusOK, &mailbus.SubscriptionResponse{
		Message: thankyouMessage,
	})

	return nil
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("ContentType", "application/json")
	w.WriteHeader(statusCode)
	//nolint:errcheck
	json.NewEncoder(w).Encode(response)
}
