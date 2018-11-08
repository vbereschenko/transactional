package transactional

import (
	"github.com/stretchr/testify/assert"
	"io"
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
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			exec := chainExecutable{name: testcase.name, steps: testcase.steps}

			err := exec.Execute(testcase.input...)

			assert.Equal(t, testcase.err, err)
		})
	}
}
