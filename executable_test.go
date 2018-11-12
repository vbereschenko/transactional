package transactional

import (
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"os"
	"testing"
)

func TestChainExecutable_Execute(t *testing.T) {
	testcases := []struct {
		name  string
		input []interface{}
		steps []step
		err   error
	}{
		{
			"regular",
			[]interface{}{"test"},
			[]step{
				&basicStep{name: "step-1", fn: func(string) {}},
			},
			nil,
		},
		{
			"regular_panic",
			[]interface{}{"test"},
			[]step{
				&basicStep{name: "step-1", fn: func(string) { panic("test") }},
			},
			failedStep,
		},
		{
			"regular_panic_error",
			[]interface{}{"test"},
			[]step{
				&basicStep{name: "step-1", fn: func(string) { panic(io.EOF) }},
			},
			io.EOF,
		},
		{
			"fallback",
			[]interface{}{"test"},
			[]step{
				&basicStep{name: "step-1", fn: func(string) {}},
				&fallbackStep{
					basicStep: basicStep{name: "step-1", fn: func() error { return nil }},
					fallback:  func(err error) {},
				},
			},
			nil,
		},
		{
			"fallback_ran",
			[]interface{}{},
			[]step{
				&basicStep{name: "step-1", fn: func() {}},
				&fallbackStep{
					basicStep: basicStep{name: "step-2", fn: func() {}, hasError: false},
					fallback:  func(err error) {},
				},
				&basicStep{name: "step-3", fn: func() error { return io.EOF }, hasError: true},
			},
			io.EOF,
		},
		{
			"fallback_double",
			[]interface{}{},
			[]step{
				&fallbackStep{
					basicStep: basicStep{name: "step-1", fn: func() {}, hasError: false},
					fallback:  func(err error) {},
				},
				&fallbackStep{
					basicStep: basicStep{name: "step-2", fn: func() error { return io.EOF }, hasError: true},
					fallback:  func(err error) {},
				},
			},
			io.EOF,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			exec := chainExecutable{name: testcase.name, steps: testcase.steps}

			err := exec.Execute(testcase.input...)

			assert.Equal(t, testcase.err, err)
		})
	}
}

func TestChainExecutable_Apply(t *testing.T) {
	l := log.New(os.Stdout, "", log.LstdFlags)
	testcases := []struct {
		expectedName, name     string
		expectedLogger, logger Logger
	}{
		{"1", "1", l, l},
		{"transaction", "", defaultLog, nil},
	}

	for _, testcase := range testcases {
		t.Run("", func(t *testing.T) {
			data := &chainExecutable{}
			config := Configuration{Name: testcase.name, Logger: testcase.logger}

			data.apply(config)

			assert.Equal(t, testcase.expectedLogger, data.logger)
			assert.Equal(t, testcase.expectedName, data.name)
		})
	}
}
