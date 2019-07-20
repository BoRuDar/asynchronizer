package asynchronizer

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func SomeJob(ctx context.Context) (Result, error) {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(5) + 1
	fmt.Println("t: ", n)

	fmt.Println(ctx.Value("test"))

	var s = fmt.Sprint("name ", n)
	//if n%2 == 0 {
	//	return Result{fmt.Sprint("name ", n), s}, fmt.Errorf("test err")
	//}
	time.Sleep(time.Duration(n) * time.Second)

	return Result{s, s}, nil
}

func TestName(t *testing.T) {
	ctx, _ := context.WithTimeout(context.TODO(), 5*time.Second)

	ctx = context.WithValue(ctx, "test", "test")

	res, err := ExecuteAsync(ctx, SomeJob, SomeJob, SomeJob, SomeJob)
	fmt.Println(res, err)

	var ss []string
	for _, r := range res {
		s, ok := r.Result.(string)
		if !ok {
			continue
		}
		ss = append(ss, s)
	}
	t.Log(strings.Join(ss, "-"))
}
