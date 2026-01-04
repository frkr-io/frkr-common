package errors

import "fmt"

// Common error types for frkr services

// ErrStreamNotFound indicates a stream was not found
type ErrStreamNotFound struct {
	StreamID string
}

func (e *ErrStreamNotFound) Error() string {
	return fmt.Sprintf("stream not found: %s", e.StreamID)
}

// ErrUnauthorized indicates authentication failed
type ErrUnauthorized struct {
	Reason string
}

func (e *ErrUnauthorized) Error() string {
	return fmt.Sprintf("unauthorized: %s", e.Reason)
}

// ErrForbidden indicates authorization failed
type ErrForbidden struct {
	Reason string
}

func (e *ErrForbidden) Error() string {
	return fmt.Sprintf("forbidden: %s", e.Reason)
}

// ErrEncryptionFailed indicates encryption/decryption failed
type ErrEncryptionFailed struct {
	Reason string
}

func (e *ErrEncryptionFailed) Error() string {
	return fmt.Sprintf("encryption failed: %s", e.Reason)
}

// ErrInvalidRequest indicates an invalid request
type ErrInvalidRequest struct {
	Reason string
}

func (e *ErrInvalidRequest) Error() string {
	return fmt.Sprintf("invalid request: %s", e.Reason)
}

