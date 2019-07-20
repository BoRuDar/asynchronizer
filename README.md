# asynchronizer

Run functions asynchronously, hides all magic. 
*100% coverage.*

## Usage

Your functions should have next signature: 
```go
    func job(ctx context.Context) (Result, error)
```

Than just put your functions and `context.Context` inside `ExecuteAsync()`:
```go
    results, err := ExecuteAsync(ctx, job1, job2)
```
