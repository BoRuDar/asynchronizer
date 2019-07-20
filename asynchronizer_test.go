// go:go test -race -v -cover -count 1 .
package asynchronizer

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

var (
	testErr = errors.New("test err")
)

func testJob1(_ context.Context) (Result, error) {
	return Result{"testJob1", "testJob1"}, nil
}

func testJob2(_ context.Context) (Result, error) {
	return Result{"testJob2", ""}, testErr
}

func testJob3(_ context.Context) (Result, error) {
	time.Sleep(10 * time.Millisecond)
	return Result{"testJob3", ""}, nil
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
	ctxWithTimeout, _ := context.WithTimeout(context.TODO(), 10*time.Millisecond)

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
			ctx:         context.TODO(),
			jobs:        nil,
			expectedErr: ErrNothingToExecute,
		},

		{
			name:        "custom error",
			ctx:         ctxWithTimeout,
			jobs:        []call{testJob1, testJob2},
			expectedErr: testErr,
		},

		{
			name:        "ctx timeout error",
			ctx:         ctxWithTimeout,
			jobs:        []call{testJob1, testJob1, testJob3},
			expectedErr: context.DeadlineExceeded,
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

func BenchmarkExecuteAsync(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, _ := context.WithTimeout(context.TODO(), 5*time.Millisecond)

		b.StartTimer()
		_, _ = ExecuteAsync(ctx, testJob1, testJob1)
		b.StopTimer()
	}
}
