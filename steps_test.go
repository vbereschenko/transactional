package transactional

import (
	"github.com/stretchr/testify/assert"
	"io"
	"reflect"
	"testing"
	"time"
)

func TestFallbackStep_Validate(t *testing.T) {
	testcases := []struct {
		name string
		step fallbackStep
		err  error
	}{
		{
			"simple",
			fallbackStep{
				basicStep: basicStep{fn: func() {}, hasError: true},
				fallback:  func(err error) {},
			},
			nil,
		},
		{
			"empty",
			fallbackStep{
				basicStep: basicStep{fn: func() {}},
				fallback:  func(err error) {},
			},
			nil,
		},
		{
			"with_data",
			fallbackStep{
				basicStep: basicStep{fn: func(field string) {}},
				fallback:  func(err error, field string) {},
			},
			nil,
		},
		{
			"invalid_fallback",
			fallbackStep{
				basicStep: basicStep{fn: func(field string) {}},
				fallback:  func(field string) {},
			},
			invalidFirstArg,
		},
		{
			"invalid_fallback",
			fallbackStep{
				basicStep: basicStep{fn: func(field string) {}},
				fallback:  func(field string) {},
			},
			invalidFirstArg,
		},
		{
			"invalid_fallback",
			fallbackStep{
				basicStep: basicStep{fn: func(field string) {}},
				fallback:  func() {},
			},
			errorHandlerRequired,
		},
		{
			"count_mismatch",
			fallbackStep{
				basicStep: basicStep{fn: func(field string) {}},
				fallback:  func(err error, field string, field2 string) {},
			},
			invalidStepFallback,
		},
		{
			"types_mismatch",
			fallbackStep{
				basicStep: basicStep{fn: func(field int) {}},
				fallback:  func(err error, field string) {},
			},
			typesMismatch,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.err, testcase.step.validate())
		})
	}
}

func TestRepeatingStep_Call(t *testing.T) {
	testcases := []struct {
		name string
		step repeatingStep
		err  error
	}{
		{
			"success",
			repeatingStep{basicStep: basicStep{fn: func() {}}, interval: time.Millisecond * 300, repeat: 5},
			nil,
		},
		{
			"error",
			repeatingStep{basicStep: basicStep{fn: func() error { return io.EOF }, hasError: true}, interval: time.Millisecond * 5, repeat: 2},
			io.EOF,
		},
		{
			"success_error",
			repeatingStep{basicStep: basicStep{fn: func() error { return nil }, hasError: true}, interval: time.Millisecond * 5, repeat: 2},
			nil,
		},
		{
			"no_call",
			repeatingStep{basicStep: basicStep{fn: func() error { return io.EOF }, hasError: true}, interval: time.Millisecond * 5, repeat: 0},
			nil,
		},
		{
			"once",
			repeatingStep{basicStep: basicStep{fn: func() error { return io.EOF }, hasError: true}, interval: time.Millisecond * 5, repeat: 1},
			io.EOF,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			_, err := testcase.step.call([]reflect.Value{})
			assert.Equal(t, testcase.err, err)
		})
	}
}

func TestRepeatingStep_Validate(t *testing.T) {
	testcases := []struct {
		name string
		step repeatingStep
		err  error
	}{
		{
			"ok",
			repeatingStep{
				basicStep: basicStep{fn: func() error { return nil }, hasError: true},
				repeat:    1,
			},
			nil,
		},
		{
			"without_err",
			repeatingStep{
				basicStep: basicStep{fn: func() {}, hasError: false},
				repeat:    1,
			},
			invalidFirstArg,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.err, testcase.step.validate())
		})
	}
}
