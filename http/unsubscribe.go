package http

import (
	"net/http"

	"github.com/quantonganh/mailbus/pkg/hash"
)

func (s *Server) unsubscribeHandler(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	email := query.Get("email")
	hashValue := query.Get("hash")
	expectedHash, err := hash.ComputeHmac256(email, s.NewsletterService.GetHMACSecret())
	if err != nil {
		return err
	}

	if hashValue == expectedHash {
		if err := s.SubscriptionService.Unsubscribe(email); err != nil {
			return err
		}

		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	return nil
}
