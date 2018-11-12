package transactional

import (
	"errors"
	"log"
	"reflect"
	"time"
)

var (
	errInterface = reflect.TypeOf((*error)(nil)).Elem()
	failedStep   = errors.New("step failed")

	errorHandlerRequired = errors.New("error handler is required")
	invalidFirstArg      = errors.New("handling function returns error, so it should be handled")
	invalidStepFallback  = errors.New("step number of arguments don't match in fn and fallback")
	typesMismatch        = errors.New("types don't match")
)

type step interface {
	getName() string
	call(values []reflect.Value) ([]reflect.Value, error)
	validate() error
}

type basicStep struct {
	name          string
	fn            interface{}
	hasError      bool
	errorPosition int
}

func (s basicStep) getName() string {
	return s.name
}

func (s basicStep) call(values []reflect.Value) (returnValues []reflect.Value, err error) {
	defer func() {
		if cause := recover(); cause != nil {
			if causeErr, is := cause.(error); is {
				err = causeErr
			} else {
				err = failedStep
			}
		}
	}()
	inputValues := reflect.ValueOf(s.fn).Call(values)

	if s.hasError && !inputValues[s.errorPosition].IsNil() {
		return nil, inputValues[s.errorPosition].Interface().(error)
	}
	if s.hasError {
		inputValues = inputValues[:len(inputValues)-1]
	}

	return inputValues, nil
}

func (s basicStep) validate() error {
	return nil
}

type fallbackStep struct {
	basicStep
	fallback interface{}
}

func (s fallbackStep) validate() error {
	rFallback := reflect.TypeOf(s.fallback)
	if rFallback.NumIn() == 0 {
		return errorHandlerRequired
	}
	if !rFallback.In(0).Implements(errInterface) {
		return invalidFirstArg
	}
	rFn := reflect.TypeOf(s.fn)

	if rFn.NumIn() != (rFallback.NumIn() - 1) {
		return invalidStepFallback
	}

	for i := 0; i < rFn.NumIn(); i++ {
		if rFn.In(i) != rFallback.In(i+1) {
			return typesMismatch
		}
	}

	return nil
}

type repeatingStep struct {
	basicStep
	interval time.Duration
	repeat   int
}

func (s repeatingStep) call(values []reflect.Value) ([]reflect.Value, error) {
	var err error
	var results []reflect.Value
	for i := 1; i <= s.repeat; i++ {
		results, err = s.basicStep.call(values)
		if err != nil {
			log.Printf("step [%s] try %d of %d, cause %e", s.name, i, s.repeat, err)
			<-time.After(s.interval)
			continue
		} else {
			return results, nil
		}

	}

	return nil, err
}

func (s repeatingStep) validate() error {
	if err := s.basicStep.validate(); err != nil {
		return err
	}

	if !s.hasError {
		return invalidFirstArg
	}

	return nil
}
