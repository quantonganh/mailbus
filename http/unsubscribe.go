package http

import (
	"net/http"

	"github.com/quantonganh/mailbus/pkg/hash"
)

const (
	unsubscribeMessage        = "Unsubscribed"
	invalidUnsubscribeMessage = "Either email or hash is invalid."
)

func (s *Server) unsubscribeHandler(w http.ResponseWriter, r *http.Request) error {
	var response struct {
		Message string `json:"message"`
	}

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

		response.Message = unsubscribeMessage
		writeJSONResponse(w, http.StatusOK, response)
	} else {
		response.Message = invalidUnsubscribeMessage
		writeJSONResponse(w, http.StatusBadRequest, response)
	}

	return nil
}
