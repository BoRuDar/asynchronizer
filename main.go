package asynchronizer

import (
	"context"
	"errors"
)

var (
	ErrNothingToExecute = errors.New("nothing to execute")
)

type call func(ctx context.Context) (Result, error)

type Result struct {
	Name   string
	Result interface{}
}

func ExecuteAsync(ctx context.Context, fn ...call) ([]Result, error) {
	if len(fn) == 0 {
		return nil, ErrNothingToExecute
	}

	ctx, cancel := context.WithCancel(ctx)
	var (
		counter int
		jobs    = len(fn)
		resCh   = make(chan Result, jobs)
		errCH   = make(chan error, jobs)
		results = make([]Result, 0, jobs)
	)
	defer func() {
		close(resCh)
		close(errCH)
	}()

	for _, f := range fn {
		go func(f call) {
			r, err := f(ctx)
			if err != nil {
				cancel()
				errCH <- err
				return
			}

			select {
			case <-ctx.Done(): // do not push results to the channel
			default:
				resCh <- r
			}
		}(f)
	}

	for {
		select {
		case res := <-resCh:
			results = append(results, res)
			counter++

			if counter == jobs {
				return results, nil
			}

		case err := <-errCH: // cancel in case of an error, no need to wait
			cancel()
			return nil, err

		case <-ctx.Done(): // global cancellation or timeout
			return nil, ctx.Err()
		}
	}
}
