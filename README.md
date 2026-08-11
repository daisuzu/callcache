# callcache

[![Go Reference](https://pkg.go.dev/badge/github.com/daisuzu/callcache.svg)](https://pkg.go.dev/github.com/daisuzu/callcache)
[![test](https://github.com/daisuzu/callcache/actions/workflows/test.yaml/badge.svg?branch=master)](https://github.com/daisuzu/callcache/actions/workflows/test.yaml)

Package callcache provides a duplicate call suppression mechanism with cache.

## Usage

```go
var dispatcher = callcache.NewDispatcher[string](1*time.Minute, 10*time.Second)

func Get(ctx context.Context, key string) (string, error) {
	return dispatcher.Do(key, func() (string, error) {
		return fetch(ctx, key)
	})
}
```
