package easyerr

import (
	"errors"
	"fmt"

	"github.com/k0kubun/pp"
	"github.com/ztrue/tracerr"
)

var ScopeInfo = true

var (
	errNotImpl       = errors.New("err not implemented")
	errNotFound      = errors.New("err not found")
	errNotSupported  = errors.New("err not supported")
	errNotAuthorized = errors.New("err not Authorized")
)

func NotImpl() error {
	if !ScopeInfo {
		return errNotImpl
	}
	return tracerr.Wrap(errNotImpl)
}

func NotFound(message string) error {
	if !ScopeInfo {
		return errNotFound
	}
	err := fmt.Errorf("%s: %w", message, errNotFound)
	return tracerr.Wrap(err)
}

func NotSupported() error {
	if !ScopeInfo {
		return errNotSupported
	}

	return tracerr.Wrap(errNotSupported)
}
func NotAuthorized() error {
	if !ScopeInfo {
		return errNotAuthorized
	}

	return tracerr.Wrap(errNotAuthorized)
}

func Error(err string) error {
	if !ScopeInfo {
		return errors.New(err)
	}

	return tracerr.New(err)
}

func Wrap(message string, err error) error {
	err = fmt.Errorf("%s: %w", message, err)
	pp.Println(err)
	if !ScopeInfo {
		return err
	}
	return tracerr.Wrap(err)
}
