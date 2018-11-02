package transactional

import (
	"log"
	"reflect"
	"time"
)

func (t Executable) Execute(input ...interface{}) error {
	inputValues := make([]reflect.Value, len(input))
	for i, inputValue := range input {
		inputValues[i] = reflect.ValueOf(inputValue)
	}
	var executionStart time.Time
	var err error
	for position, step := range t.steps {
		t.values[position] = inputValues
		executionStart = time.Now()
		inputValues, err = step.call(inputValues)
		log.Printf("step [%s] time %v", step.getName(), time.Since(executionStart))
		if err != nil {
			log.Printf("fallbackStep [%s] failed", step.getName())
			t.fallback(reflect.ValueOf(err), position)

			return err
		}
	}

	return nil
}

func (t Executable) fallback(err reflect.Value, position int) {
	for i:=position; i>=0; i-- {
		if fallbackStep, correct := t.steps[i].(fallbackStep); correct {
			log.Printf("fallbackStep [%s] rolled back", t.steps[i].getName())
			reflect.ValueOf(fallbackStep.fallback).Call(append(append([]reflect.Value{err}, t.values[i]...)))
		}
	}
}

type Executable struct {
	steps []step
	values [][]reflect.Value
}
