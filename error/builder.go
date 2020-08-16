package error

import (
	"github.com/getsentry/sentry-go"
	"github.com/google/logger"
)

type spotifeteErrorBuilder interface {
	WithMessage(message string) spotifeteErrorBuilder
	WithCause(cause error) spotifeteErrorBuilder
	Build() error

	getDefaultMessage() string
	getHttpStatus() int
	shouldShowMessageToUser() bool
	shouldShowCauseToUser() bool
	shouldLogError() bool
}

type baseErrorBuilder struct {
	spotifeteErrorBuilder

	message *string
	cause error
}

func (e baseErrorBuilder) WithMessage(message string) spotifeteErrorBuilder {
	e.message = &message
	return e
}

func (e baseErrorBuilder) WithCause(cause error) spotifeteErrorBuilder {
	e.cause = cause
	return e
}

func (e baseErrorBuilder) Build() error {
	e.logErrorIfNeccessary()

	return spotifeteError{
		message:    e.buildUserErrorMessage(),
		httpStatus: e.getHttpStatus(),
	}
}


func (e baseErrorBuilder) logErrorIfNeccessary() {
	if e.shouldLogError() {
		e.logError()
	}
}

func (e baseErrorBuilder) logError() {
	if e.cause != nil {
		sentry.CaptureException(e.cause)
		logger.ErrorDepth(999999, e.cause)
	} else if e.message != nil {
		sentry.CaptureMessage(*e.message)
		logger.ErrorDepth(999999,e.message)
	}
}

func (e baseErrorBuilder) buildUserErrorMessage() (errorMessage string) {
	if e.shouldShowMessageToUser() {
		errorMessage = e.getMessageOrDefaultIfEmpty()
	} else {
		errorMessage = e.getDefaultMessage()
	}

	if e.shouldShowCauseToUser() {
		errorMessage += e.buildCauseText()
	}

	return errorMessage
}

func (e baseErrorBuilder) getMessageOrDefaultIfEmpty() (message string) {
	if e.message != nil {
		message = *e.message
	} else {
		message = e.getDefaultMessage()
	}

	return message
}

func (e baseErrorBuilder) buildCauseText() (causeText string) {
	if e.cause != nil {
		causeText += "\n"
		causeText += "Caused by: "
		causeText += e.cause.Error()
	}

	return causeText
}

func (e baseErrorBuilder) getDefaultMessage() string {
	return "An error occurred."
}

func (e baseErrorBuilder) shouldLogError() bool {
	return true
}
