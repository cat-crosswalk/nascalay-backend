package ws

import "errors"

var (
	errNilBody          = errors.New("body is nil")
	errUnAuthorized     = errors.New("unauthorized")
	errUnknownEventType = errors.New("unknown event type")
	errWrongPhase       = errors.New("wrong phase")
)
