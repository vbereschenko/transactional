package transactional

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTransaction_Step(t *testing.T) {
	testcases := []struct {
		name          string
		hasError      bool
		errorPosition int
		fn            interface{}
	}{
		{
			"regular",
			false,
			0,
			func() {},
		},
		{
			"errored",
			true,
			0,
			func() error { return nil },
		},
		{
			"data_with_error",
			true,
			2,
			func() (string, int, error) { return "", 1, nil },
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			transaction := Transaction{}
			transaction.Step(testcase.name, testcase.fn)
			step := transaction.steps[0].(basicStep)
			assert.Equal(t, testcase.hasError, step.hasError)
			assert.Equal(t, testcase.errorPosition, step.errorPosition)
			assert.NotNil(t, step.fn)
		})
	}
}

func TestTransaction_FallbackStep(t *testing.T) {
	testcases := []struct {
		name          string
		hasError      bool
		errorPosition int
		fn            interface{}
		fallback      interface{}
	}{
		{
			"regular",
			false,
			0,
			func() {},
			func() {},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			transaction := Transaction{}
			transaction.FallbackStep(testcase.name, testcase.fn, testcase.fallback)
			step := transaction.steps[0].(fallbackStep)
			assert.Equal(t, testcase.hasError, step.hasError)
			assert.Equal(t, testcase.errorPosition, step.errorPosition)
			assert.NotNil(t, step.fn)
		})
	}
}

func TestTransaction_RepeatingStep(t *testing.T) {
	testcases := []struct {
		name          string
		hasError      bool
		errorPosition int
		fn            interface{}
	}{
		{
			"regular",
			false,
			0,
			func() {},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			transaction := Transaction{}
			transaction.RepeatingStep(testcase.name, testcase.fn, 1, time.Millisecond)
			assert.IsType(t, repeatingStep{}, transaction.steps[0])
			step := transaction.steps[0].(repeatingStep)
			assert.Equal(t, testcase.hasError, step.hasError)
			assert.Equal(t, testcase.errorPosition, step.errorPosition)
			assert.NotNil(t, step.fn)
		})
	}
}

func TestTransaction_Build(t *testing.T) {
	testcases := []struct {
		name          string
		steps         []step
		errorExpected bool
	}{
		{
			"correct",
			[]step{repeatingStep{basicStep: basicStep{fn: func() {}, hasError: true}}},
			false,
		},
		{
			"validation",
			[]step{repeatingStep{basicStep: basicStep{fn: func() {}, hasError: false}}},
			true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			chain := Transaction{steps: testcase.steps}
			_, err := chain.Build()
			_ = testcase.errorExpected && assert.Error(t, err)
		})
	}
}
