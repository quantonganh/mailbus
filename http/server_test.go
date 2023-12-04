package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/asdine/storm/v3"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/mailbus"
	"github.com/quantonganh/mailbus/mock"
	"github.com/quantonganh/mailbus/pkg/hash"
)

var (
	cfg *mailbus.Config
	s   *Server
)

func TestMain(m *testing.M) {
	viper.SetConfigType("yaml")
	var yamlConfig = []byte(`
templates:
  dir: html/templates

newsletter:
  hmac:
    secret: da02e221bc331c9875c5e1299fa8d765
`)
	if err := viper.ReadConfig(bytes.NewBuffer(yamlConfig)); err != nil {
		log.Fatal(err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}

	var err error
	s, err = NewServer()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func TestSubscribeHandler(t *testing.T) {
	t.Parallel()

	email := "foo@gmail.com"
	token := uuid.NewV4().String()

	subscribe := &mailbus.Subscription{}
	subscribeService := new(mock.SubscriptionService)
	subscribeService.On("FindByEmail", email).Return(subscribe, storm.ErrNotFound)
	subscribeService.On("Insert", mailbus.NewSubscription(email, token, mailbus.StatusPending)).Return(nil)

	smtpService := new(mock.NewsletterService)
	smtpService.On("SendConfirmationEmail", email, token).Return(nil)
	smtpService.On("GenerateNewUUID").Return(token)

	s.SubscriptionService = subscribeService
	s.NewsletterService = smtpService

	subscriptionReq := &mailbus.SubscriptionRequest{
		Email: email,
	}
	data, err := json.Marshal(subscriptionReq)
	assert.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, "/subscribe", bytes.NewReader(data))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, w.Code)
	var subscriptionResp *mailbus.SubscriptionResponse
	err = json.NewDecoder(resp.Body).Decode(&subscriptionResp)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(confirmationMessage, email), subscriptionResp.Message)
}

func TestConfirmHandler(t *testing.T) {
	email := "foo@gmail.com"
	token := uuid.NewV4().String()

	subscribe := mailbus.NewSubscription(email, token, blog.StatusPending)
	subscribeService := new(mock.SubscriptionService)
	subscribeService.On("Subscribe", token).Return(nil)
	subscribeService.On("FindByToken", token).Return(subscribe, nil)

	smtpService := new(mock.NewsletterService)
	smtpService.On("SendThankYouEmail", email).Return(nil)

	s.SubscriptionService = subscribeService
	s.NewsletterService = smtpService

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/subscribe/confirm?token=%s", token), nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, w.Code)
	var subscriptionResp *mailbus.SubscriptionResponse
	err = json.NewDecoder(resp.Body).Decode(&subscriptionResp)
	assert.Equal(t, thankyouMessage, subscriptionResp.Message)
}

func TestUnsubscribeHandler(t *testing.T) {
	email := "foo@gmail.com"
	secret := cfg.Newsletter.HMAC.Secret
	hashValue, err := hash.ComputeHmac256(email, secret)
	require.NoError(t, err)

	subscriptionService := new(mock.SubscriptionService)
	subscriptionService.On("Unsubscribe", email).Return(nil)

	s.SubscriptionService = subscriptionService

	newsletterService := new(mock.NewsletterService)
	newsletterService.On("GetHMACSecret").Return(secret)
	s.NewsletterService = newsletterService

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/unsubscribe?email=%s&hash=%s", email, hashValue), nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, w.Code)
	var subscriptionResp *mailbus.SubscriptionResponse
	err = json.NewDecoder(resp.Body).Decode(&subscriptionResp)
	assert.Equal(t, unsubscribeMessage, subscriptionResp.Message)
}
