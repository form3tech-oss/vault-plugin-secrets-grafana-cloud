package secretsengine

import "fmt"

type InvalidConfigurationError struct {
	Msg string
	Err error
}

func NewInvalidConfigurationError(msg string, err error) *InvalidConfigurationError {
	return &InvalidConfigurationError{
		Msg: msg,
		Err: err,
	}
}

func (e *InvalidConfigurationError) Error() string {
	return fmt.Sprintf("invalid configuration: %s", e.Msg)
}

func (e *InvalidConfigurationError) Unwrap() error {
	return e.Err
}

type InternalError struct {
	Msg string
	Err error
}

func NewInternalError(msg string, err error) *InternalError {
	return &InternalError{
		Msg: msg,
		Err: err,
	}
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("internal error: %s", e.Msg)
}

func (e *InternalError) Unwrap() error {
	return e.Err
}
