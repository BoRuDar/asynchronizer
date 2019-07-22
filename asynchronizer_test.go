package asynchronizer

import (
	"context"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var (
	testErr = errors.New("test err")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func testJob1(_ context.Context) (Result, error) {
	return Result{"testJob1", "testJob1"}, nil
}

func testJob2(_ context.Context) (Result, error) {
	return Result{"testJob2", ""}, testErr
}

func testJob3(_ context.Context) (Result, error) {
	time.Sleep(50 * time.Millisecond)
	return Result{"testJob3", ""}, nil
}

type job struct {
	id  string
	url string
}

func (j job) call(_ context.Context) (r Result, err error) {
	r.Identifier = j.id

	resp, err := http.Get(j.url)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return r, err
	}

	r.Result = string(body)
	return
}

func testHTTPSrv() *httptest.Server {
	r := http.NewServeMux()
	r.HandleFunc("/one", func(w http.ResponseWriter, _ *http.Request) {
		randSleep()
		w.Write([]byte("one"))
	})
	r.HandleFunc("/two", func(w http.ResponseWriter, _ *http.Request) {
		randSleep()
		w.Write([]byte("two"))
	})

	return httptest.NewServer(r)
}

func randSleep() {
	time.Sleep(time.Duration(1+rand.Intn(10)) * time.Millisecond)
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
			ctx:         context.TODO(),
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

	t.Parallel()
	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if _, gotErr := ExecuteAsync(test.ctx, test.jobs...); gotErr != test.expectedErr {
				t.Fatalf(`expected error: [%v], but got: [%v]`, test.expectedErr, gotErr)
			}
		})
	}
}

func TestHttpCalls(t *testing.T) {
	srv := testHTTPSrv()
	defer srv.Close()

	j1 := job{id: "one", url: srv.URL + "/one"}
	j2 := job{id: "two", url: srv.URL + "/two"}

	if _, err := ExecuteAsync(context.TODO(), j1.call, j2.call); err != nil {
		t.Fatal("unexpected err: ", err)
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
