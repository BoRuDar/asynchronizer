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
