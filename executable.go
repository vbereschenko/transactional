package transactional

import (
	"log"
	"reflect"
	"time"
)

// executable transaction interface
type Executable interface {
	// to run transaction, first function will receive arguments provided to this func
	Execute(input ...interface{}) error
}

type chainExecutable struct {
	name  string
	steps []step
}

func (t chainExecutable) Execute(input ...interface{}) error {
	inputValues := make([]reflect.Value, len(input))
	for i, inputValue := range input {
		inputValues[i] = reflect.ValueOf(inputValue)
	}
	var executionStart time.Time
	var err error
	var values = make([][]reflect.Value, len(t.steps))
	for position, step := range t.steps {
		values[position] = inputValues
		executionStart = time.Now()
		inputValues, err = step.call(inputValues)
		log.Printf("[%s] step [%s] time %v", t.name, step.getName(), time.Since(executionStart))
		if err != nil {
			log.Printf("[%s] step [%s] failed, trying to rollback changes", t.name, step.getName())
			t.fallback(values, reflect.ValueOf(err), position)

			return err
		}
	}

	return nil
}

func (t chainExecutable) fallback(values [][]reflect.Value, err reflect.Value, position int) {
	for i := position; i >= 0; i-- {
		if fallbackStep, correct := t.steps[i].(*fallbackStep); correct {
			log.Printf("[%s] fallbackStep [%s] rolled back", t.name, t.steps[i].getName())
			reflect.ValueOf(fallbackStep.fallback).Call(append(append([]reflect.Value{err}, values[i]...)))
		}
	}
}
