package transactional

import (
	"log"
	"os"
	"reflect"
	"time"
)

var (
	defaultLog = log.New(os.Stdout, "[transaction] ", log.LstdFlags)
)

// executable transaction interface
type Executable interface {
	// to run transaction, first function will receive arguments provided to this func
	Execute(input ...interface{}) error
}

type chainExecutable struct {
	name   string
	steps  []step
	logger Logger
}

// Execute default chain
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
		log.Printf("[%s] -> [%s] calling", t.name, step.getName())
		inputValues, err = step.call(inputValues)
		log.Printf("[%s] -> [%s] time %v", t.name, step.getName(), time.Since(executionStart))
		if err != nil {
			log.Printf("[%s] -> [%s] failed, trying to rollback changes: %e", t.name, step.getName(), err)
			t.fallback(values, reflect.ValueOf(err), position)

			return err
		}
	}

	return nil
}

func (t chainExecutable) fallback(values [][]reflect.Value, err reflect.Value, position int) {
	for i := position; i >= 0; i-- {
		if fallbackStep, correct := t.steps[i].(*fallbackStep); correct {
			log.Printf("[%s] -> [%s] rolled back", t.name, t.steps[i].getName())
			reflect.ValueOf(fallbackStep.fallback).Call(append(append([]reflect.Value{err}, values[i]...)))
		}
	}
}

func (t *chainExecutable) apply(config Configuration) *chainExecutable {
	t.logger = config.Logger
	if t.logger == nil {
		t.logger = defaultLog
	}

	t.name = config.Name
	if t.name == "" {
		t.name = "transaction"
	}

	return t
}
