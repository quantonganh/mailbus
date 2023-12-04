// Code generated by mockery v2.33.0. DO NOT EDIT.

package mock

import mock "github.com/stretchr/testify/mock"

// NewsletterService is an autogenerated mock type for the NewsletterService type
type NewsletterService struct {
	mock.Mock
}

// GenerateNewUUID provides a mock function with given fields:
func (_m *NewsletterService) GenerateNewUUID() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetHMACSecret provides a mock function with given fields:
func (_m *NewsletterService) GetHMACSecret() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// SendConfirmationEmail provides a mock function with given fields: to, token
func (_m *NewsletterService) SendConfirmationEmail(to string, token string) error {
	ret := _m.Called(to, token)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(to, token)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendNewsletter provides a mock function with given fields: content
func (_m *NewsletterService) SendNewsletter(content string) {
	_m.Called(content)
}

// SendThankYouEmail provides a mock function with given fields: to
func (_m *NewsletterService) SendThankYouEmail(to string) error {
	ret := _m.Called(to)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(to)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Stop provides a mock function with given fields:
func (_m *NewsletterService) Stop() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewNewsletterService creates a new instance of NewsletterService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewNewsletterService(t interface {
	mock.TestingT
	Cleanup(func())
}) *NewsletterService {
	mock := &NewsletterService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}