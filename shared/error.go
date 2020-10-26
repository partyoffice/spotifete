package shared

import (
	"github.com/getsentry/sentry-go"
	"github.com/google/logger"
	"net/http"
)

type SpotifeteError struct {
	MessageForUser string
	HttpStatus     int
}

// Creates an error struct with the supplied message, cause and http status code
// The message and cause will be logged
//
// This is useful for most cases, where the reason might be interesting for the user
func NewError(message string, cause error, httpStatus int) *SpotifeteError {
	defer logErrorAsync(message, cause, 1)

	return &SpotifeteError{
		MessageForUser: buildMessageWithCause(message, cause),
		HttpStatus:     httpStatus,
	}
}

// Creates an error struct with the supplied message and http status 400 (Bad Request)
// Nothing will be logged
//
// This is useful for user errors like missing parameters
func NewUserError(message string) *SpotifeteError {
	return &SpotifeteError{
		MessageForUser: message,
		HttpStatus:     http.StatusBadRequest,
	}
}

// Creates an error struct with a standard message and http status 500 (Internal Server Error)
// The message and cause will be logged
//
// This is useful if you don't want the user to see the details of the error
func NewInternalError(message string, cause error) *SpotifeteError {
	defer logErrorAsync(message, cause, 1)

	return &SpotifeteError{
		MessageForUser: "An error has occurred.",
		HttpStatus:     http.StatusInternalServerError,
	}
}

func logErrorAsync(message string, cause error, logDepth int) {
	logError(message, cause, logDepth+1)
}

func logError(message string, cause error, logDepth int) {
	logger.ErrorDepth(logDepth+1, buildMessageWithCause(message, cause))
	sentry.CaptureMessage(message)
	sentry.CaptureException(cause)
}

func buildMessageWithCause(message string, cause error) string {
	if cause == nil {
		return message
	} else {
		return message + "\ncaused by: " + cause.Error()
	}
}
