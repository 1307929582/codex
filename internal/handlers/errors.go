package handlers

import "errors"

type userError struct {
	msg string
}

func (e userError) Error() string {
	return e.msg
}

func newUserError(msg string) error {
	return userError{msg: msg}
}

func isUserError(err error) bool {
	var ue userError
	return errors.As(err, &ue)
}
