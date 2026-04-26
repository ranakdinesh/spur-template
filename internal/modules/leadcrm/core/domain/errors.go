package domain

import "errors"

var (
	ErrLeadNotFound      = errors.New("lead not found")
	ErrContactNotFound   = errors.New("contact not found")
	ErrAccountNotFound   = errors.New("account not found")
	ErrWebFormNotFound   = errors.New("web form not found")
	ErrInvalidSubmission = errors.New("invalid submission")
)
