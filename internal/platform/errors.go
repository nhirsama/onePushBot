package platform

import "errors"

var (
	ErrBusClosed          = errors.New("platform bus closed")
	ErrSubscriptionClosed = errors.New("platform subscription closed")
	ErrClientNil          = errors.New("platform client is nil")
	ErrClientExists       = errors.New("platform client already registered")
	ErrClientInvalid      = errors.New("platform client is invalid")
	ErrHubStarted         = errors.New("platform hub already started")
)
