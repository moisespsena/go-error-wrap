package errwrap

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type ErrorWrapperInterface interface {
	error
	Err() error
	First() error
	Prev() ErrorWrapperInterface
	List() []error
	Each(func(err error) error) error
	EachType(cb func(typ reflect.Type, err error) error) (err error)
	Is(err error) bool
}

type Wrapper struct {
	err  error
	prev ErrorWrapperInterface
}

func (w Wrapper) Error() string {
	var parts []string
	w.Each(func(err error) error {
		parts = append(parts, err.Error())
		return nil
	})
	return strings.Join(parts, " <-- ")
}

func (w Wrapper) Err() (err error) {
	return w.err
}

func (w Wrapper) First() (err error) {
	w.Each(func(e error) error {
		err = e
		return nil
	})
	return
}

func (w Wrapper) Prev() ErrorWrapperInterface {
	return w.prev
}

func (w Wrapper) List() (errors []error) {
	var wr ErrorWrapperInterface = w
	for wr != nil {
		errors = append(errors, wr.Err())
		wr = wr.Prev()
	}
	return
}

func (w Wrapper) Each(cb func(err error) error) (err error) {
	var wr ErrorWrapperInterface = w
	for wr != nil {
		err = cb(wr.Err())
		if err != nil {
			return err
		}
		wr = wr.Prev()
	}
	return
}

func (w Wrapper) EachType(cb func(typ reflect.Type, err error) error) (err error) {
	var wr ErrorWrapperInterface = w
	for wr != nil {
		err = cb(TypeOf(wr.Err()), wr.Err())
		if err != nil {
			return err
		}
		wr = wr.Prev()
	}
	return
}

func (w Wrapper) Is(err error) (is bool) {
	w.Each(func(arg error) error {
		if err == arg {
			is = true
			return io.EOF
		}
		return nil
	})
	return
}

func Wrap(child error, self interface{}, args ...interface{}) ErrorWrapperInterface {
	if child == nil {
		return nil
	}
	if s, ok := self.(string); ok {
		if len(args) == 0 {
			self = errors.New(s)
		} else {
			self = fmt.Errorf(s, args...)
		}
	}
	if !Wrapped(child) {
		child = &Wrapper{err: child}
	}
	return &Wrapper{self.(error), child.(ErrorWrapperInterface)}
}

func Wrapped(err error) bool {
	_, ok := err.(ErrorWrapperInterface)
	return ok
}

func TypeOf(e error) (typ reflect.Type) {
	typ = reflect.TypeOf(e)
	for typ.Kind() == reflect.Ptr || typ.Kind() == reflect.Interface {
		typ = typ.Elem()
	}
	return
}
