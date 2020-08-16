package error

type spotifeteError struct {
	error

	message string
	cause error
	httpStatus int
}

func (e spotifeteError) Error() string {
	return e.message
}

func (e spotifeteError) Unwrap() error {
	return e.cause
}

func (e spotifeteError) GetHttpStatus() int {
	return e.httpStatus
}
