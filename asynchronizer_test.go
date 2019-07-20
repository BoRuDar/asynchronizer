package asynchronizer

import (
	"context"
	"strings"
	"testing"
	"time"
)

func testJob1(_ context.Context) (Result, error) {
	return Result{"testJob1", "testJob1"}, nil
}

func TestExecuteAsyncOK(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Millisecond)
	defer cancel()

	const expectedResult = `testJob1-testJob1`

	res, err := ExecuteAsync(ctx, testJob1, testJob1)
	if err != nil {
		t.Fatal(`unexpected error: `, err)
	}

	if len(res) != 2 {
		t.Fatal(`wrong number of results: `, len(res))
	}

	var ss []string
	for _, r := range res {
		s, ok := r.Result.(string)
		if ok {
			ss = append(ss, s)
		}
	}

	gotResult := strings.Join(ss, "-")
	if gotResult != expectedResult {
		t.Fatalf(`expected: [%s], but got: [%s]`, expectedResult, gotResult)
	}
}

func TestExecuteAsyncErrors(t *testing.T) {
	testCases := []struct {
		name        string
		ctx         context.Context
		jobs        []call
		expectedErr error
	}{
		{
			name:        "nil context",
			ctx:         nil,
			jobs:        []call{testJob1},
			expectedErr: ErrNilContext,
		},

		{
			name:        "no functions for the execution",
			ctx:         nil,
			jobs:        nil,
			expectedErr: ErrNothingToExecute,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if _, gotErr := ExecuteAsync(test.ctx, test.jobs...); gotErr != test.expectedErr {
				t.Fatalf(`expected error: [%v], but got: [%v]`, test.expectedErr, gotErr)
			}
		})
	}
}
