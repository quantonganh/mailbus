package http

import (
	"encoding/json"
	"net/http"

	"github.com/quantonganh/mailbus"
)

func (s *Server) sendNewsletterHandler(w http.ResponseWriter, r *http.Request) error {
	activeSubscribers, err := s.SubscriptionService.FindByStatus(mailbus.StatusActive)
	if err != nil {
		return err
	}

	var req *mailbus.EmailNewsletterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	s.NewsletterService.SendNewsletter(activeSubscribers, req.Subject, req.Body)

	return nil
}
