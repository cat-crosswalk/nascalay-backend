package ws

import "errors"

var (
	errNilBody          = errors.New("body is nil")
	errUnAuthorized     = errors.New("unauthorized")
	errUnknownEventType = errors.New("unknown event type")
	errWrongPhase       = errors.New("wrong phase")
	errNotFound         = errors.New("not found")
	errAlreadyExists    = errors.New("already exists")
	errUnknownPhase     = errors.New("unknown phase")
	errInvalidDrawCount = errors.New("invalid draw count")
	errNotEnoughMember  = errors.New("not enough member")
)
