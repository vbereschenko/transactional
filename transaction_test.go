package transactional

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
