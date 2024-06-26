package http

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"

	"github.com/quantonganh/httperror"
	"github.com/quantonganh/mailbus"
)

const (
	shutdownTimeout = 1 * time.Second
)

// Server represents HTTP server
type Server struct {
	ln     net.Listener
	server *http.Server
	router *mux.Router

	Addr   string
	Domain string

	SubscriptionService mailbus.SubscriptionService
	NewsletterService   mailbus.NewsletterService
	QueueService        mailbus.QueueService
}

// NewServer create new HTTP server
func NewServer() (*Server, error) {
	s := &Server{
		server: &http.Server{},
		router: mux.NewRouter().StrictSlash(true),
	}

	zlog := zerolog.New(os.Stdout).With().
		Timestamp().
		Logger()
	s.router.Use(hlog.NewHandler(zlog))
	s.router.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Str("form_value", r.FormValue("letters")).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	s.router.Use(hlog.UserAgentHandler("user_agent"))
	s.router.Use(hlog.RefererHandler("referer"))
	s.router.Use(httperror.RequestIDHandler("req_id"))

	sentryHandler := sentryhttp.New(sentryhttp.Options{})
	s.router.Use(sentryHandler.Handle)

	s.server.Handler = http.HandlerFunc(s.serveHTTP)

	s.router.HandleFunc("/health", s.healthCheckHandler)
	s.router.HandleFunc("/subscriptions", s.Error(s.subscriptionsHandler)).Methods(http.MethodPost)
	subRouter := s.router.PathPrefix("/subscriptions").Subrouter()
	subRouter.HandleFunc("/confirm", s.Error(s.confirmHandler))
	s.router.HandleFunc("/unsubscribe", s.Error(s.unsubscribeHandler))

	return s, nil
}

// Scheme returns scheme
func (s *Server) Scheme() string {
	if s.UseTLS() {
		return "https"
	}
	return "http"
}

// UseTLS checks if server use TLS or not
func (s *Server) UseTLS() bool {
	return s.Domain != ""
}

// Port returns server port
func (s *Server) Port() int {
	if s.ln == nil {
		return 0
	}
	return s.ln.Addr().(*net.TCPAddr).Port
}

// URL returns server URL
func (s *Server) URL() string {
	scheme, port := s.Scheme(), s.Port()

	domain := "localhost"
	if s.Domain != "" {
		domain = s.Domain
	}

	if port == 80 || port == 443 || flag.Lookup("test.v") != nil {
		return fmt.Sprintf("%s://%s", scheme, domain)
	}
	return fmt.Sprintf("%s://%s:%d", scheme, domain, s.Port())
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Open opens a connection to HTTP server
func (s *Server) Open() (err error) {
	s.ln, err = net.Listen("tcp", s.Addr)
	if err != nil {
		return errors.Errorf("failed to listen to port %s: %v", s.Addr, err)
	}

	go func() {
		_ = s.server.Serve(s.ln)
	}()

	return nil
}

func (s *Server) ConsumeAndSendNewsletter(ctx context.Context) error {
	messages, err := s.QueueService.Consume(ctx, "added-posts")
	if err != nil {
		return err
	}

	activeSubscribers, err := s.SubscriptionService.FindByStatus(mailbus.StatusActive)
	if err != nil {
		return err
	}

	for msg := range messages {
		var req *mailbus.EmailNewsletterRequest
		if err := json.NewDecoder(bytes.NewReader(msg)).Decode(&req); err != nil {
			return err
		}

		s.NewsletterService.SendNewsletter(activeSubscribers, req.Subject, req.Body)
	}

	return nil
}

// Close shutdowns HTTP server
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}
