package asynchronizer

import (
	"context"
	"errors"
	"sync"
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
		wg      = &sync.WaitGroup{}
		err     error
	)
	defer func() {
		close(resCh)
		close(errCh)
	}()

	for _, f := range fn {
		wg.Add(1)
		go func(f call) {
			defer wg.Done()
			r, err := f(ctx)
			select {
			case <-ctx.Done():
				return
			default:
				if err != nil {
					errCh <- err
					return
				}
				resCh <- r
			}
		}(f)
	}

outer:
	for {
		select {
		case res := <-resCh:
			results = append(results, res)
			counter++

			if counter == jobs {
				break outer
			}

		case e := <-errCh: // cancel in case of an error, no need to wait
			cancel()
			err = e
			break outer

		case <-ctx.Done(): // global cancellation or timeout
			err = ctx.Err()
			break outer
		}
	}

	wg.Wait()
	if err != nil {
		return nil, err
	}
	return results, nil
}
