package errors

import (
	"errors"
	"fmt"
)

type Error struct {
	ferr error
	serr error
}

func (e *Error) Unwrap() error {
	return e.serr
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v, %v", e.ferr, e.serr)
}

func Wrap(ferr, serr error) error {
	return &Error{ferr, serr}
}

func New(msg string) error {
	return errors.New(msg)
}

func Newf(f string, args ...any) error {
	return fmt.Errorf(f, args...)
}

func NewWrap(msg string, err error) error {
	return &Error{errors.New(msg), err}
}
