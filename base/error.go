package base

import "github.com/pkg/errors"

var (
	// Backend error responses
	ErrPermissionDeniedResponse  = errors.New("received permission denied response")
	ErrNotFoundResponse          = errors.New("received not found response from backend")
	ErrBackendErrorResponse      = errors.New("received internal error response form backend")
	ErrUnexpectedMessageResponse = errors.New("received response with unexpected message")

	// Storage error responses
	ErrStorageNotFound     = errors.New("received storage not found response")
	ErrStorageAccessDenied = errors.New("received storage access denied response")
	ErrStorageOtherError   = errors.New("received other storage error response")

	// Other response
	ErrUnknownTypeResponse = errors.New("received unknown type response")
)
