package asynchronizer

import (
	"context"
	"errors"
)

var (
	ErrNothingToExecute = errors.New("nothing to execute")
	ErrNilContext       = errors.New("nil context")
)

type call func(ctx context.Context) (Result, error)

type Result struct {
	Identifier string // should be unique
	Result     interface{}
}

func ExecuteAsync(ctx context.Context, fn ...call) ([]Result, error) {
	if len(fn) == 0 {
		return nil, ErrNothingToExecute
	}
	if ctx == nil {
		return nil, ErrNilContext
	}

	ctx, cancel := context.WithCancel(ctx)
	var (
		counter int
		jobs    = len(fn)
		resCh   = make(chan Result, jobs)
		errCh   = make(chan error, jobs)
		results = make([]Result, 0, jobs)
	)
	defer func() {
		close(resCh)
		close(errCh)
	}()

	for _, f := range fn {
		go func(f call) {
			r, err := f(ctx)
			if err != nil {
				errCh <- err
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

		case err := <-errCh: // cancel in case of an error, no need to wait
			cancel()
			return nil, err

		case <-ctx.Done(): // global cancellation or timeout
			return nil, ctx.Err()
		}
	}
}
