package client

type ClientError struct {
	Msg string
	Err error
}

func NewClientError(msg string, err error) *ClientError {
	return &ClientError{
		Msg: msg,
		Err: err,
	}
}

func (e *ClientError) Error() string {
	return e.Msg
}

func (e *ClientError) Unwrap() error {
	return e.Err
}
