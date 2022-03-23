package pkg

import "errors"

var (
	// Common errors.
	ErrMethodParamsInvalid = errors.New("params passed to method is invalid")

	// Protocol errors.
	ErrTotalLengthInvalid    = errors.New("CommandLength in Packet data is invalid")
	ErrCommandIDInvalid      = errors.New("CommandID in Packet data is invalid")
	ErrCommandIDNotSupported = errors.New("CommandID in Packet data is not supported")

	// Connection errors.
	ErrConnIsClosed       = errors.New("connection is closed")
	ErrReadHeaderTimeout  = errors.New("read header timeout")
	ErrReadPktBodyTimeout = errors.New("read packet body timeout")
)

type OpError struct {
	err error
	op  string
}

func NewOpError(e error, op string) *OpError {
	return &OpError{
		err: e,
		op:  op,
	}
}

func (e *OpError) Error() string {
	if e.err == nil {
		return "<nil>"
	}
	return e.op + " error: " + e.err.Error()
}

func (e *OpError) Cause() error {
	return e.err
}

func (e *OpError) Op() string {
	return e.op
}
