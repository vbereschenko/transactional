package transactional

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"
)

var (
	errInterface = reflect.TypeOf((*error)(nil)).Elem()
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

func (s basicStep) call(values []reflect.Value) ([]reflect.Value, error) {
	log.Printf("step [%s] calling", s.name)
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
	hasError := 0
	rFallback := reflect.TypeOf(s.fallback)
	if s.hasError {
		if !rFallback.In(0).Implements(errInterface) {
			return errors.New("handling function returns error, so it should be handled")
		}
		hasError++
	}
	rFn := reflect.TypeOf(s.fn)

	if rFn.NumIn() != rFallback.NumIn()-hasError {
		return fmt.Errorf("step [%s] number of arguments don't match in fn and fallback", s.name)
	}

	for i := 0; i < rFn.NumIn(); i++ {
		if rFn.In(i) != rFallback.In(i+hasError) {
			return errors.New("types don't match")
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
			log.Printf("step [%s] repeating %d of %d, cause %e", s.name, i, s.repeat, err)
			<-time.After(s.interval)
			continue
		} else {
			return results, nil
		}

	}

	return nil, err
}
