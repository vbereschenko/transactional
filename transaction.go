package transactional

import (
	"reflect"
	"time"
)

type Transaction struct {
	Name  string
	steps []step
}

func (t *Transaction) Step(name string, fn interface{}) *Transaction {
	step := basicStep{name: name, fn: fn}
	step.hasError, step.errorPosition = errorPosition(fn)

	t.steps = append(t.steps, step)

	return t
}

func (t *Transaction) FallbackStep(name string, fn, fallback interface{}) *Transaction {
	step := fallbackStep{
		basicStep: basicStep{name: name, fn: fn},
	}
	step.hasError, step.errorPosition = errorPosition(fn)
	step.fallback = fallback

	t.steps = append(t.steps, step)

	return t
}

func (t *Transaction) RepeatingStep(name string, fn interface{}, repeat int, interval time.Duration) {
	step := repeatingStep{
		basicStep: basicStep{name: name, fn: fn},
		repeat:    repeat,
		interval:  interval,
	}
	step.hasError, step.errorPosition = errorPosition(fn)

	t.steps = append(t.steps, step)
}

func (t Transaction) Build() (Executable, error) {
	for _, step := range t.steps {
		if err := step.validate(); err != nil {
			return nil, err
		}
	}

	return &chainExecutable{t.Name, t.steps}, nil
}

func errorPosition(step interface{}) (bool, int) {
	stepType := reflect.TypeOf(step)
	if stepType.NumOut() == 0 {
		return false, 0
	}
	lastType := stepType.Out(stepType.NumOut() - 1)

	return lastType.Implements(errInterface), stepType.NumOut() - 1
}