package callcache_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/daisuzu/callcache"
)

func ExampleDispatcher() {
	dispatcher := callcache.NewDispatcher[string](1*time.Minute, 10*time.Second)

	v, err := dispatcher.Do("key", func() (string, error) {
		return "example", nil
	})

	fmt.Println(v, err)
	// Output:
	// example <nil>
}

func ExampleNewDispatcher_expiration() {
	dispatcher := callcache.NewDispatcher[int](1*time.Nanosecond, 0)

	wg := sync.WaitGroup{}
	results := make([]any, 3)
	for i := range results {
		wg.Add(1)
		results[i], _ = dispatcher.Do("key", func() (int, error) {
			defer wg.Done()
			fmt.Printf("Do: #%d\n", i+1)
			return i + 1, nil
		})
		wg.Wait()
		time.Sleep(1 * time.Nanosecond)
	}

	for _, v := range results {
		fmt.Println(v)
	}
	// Output:
	// Do: #1
	// Do: #2
	// Do: #3
	// 1
	// 2
	// 3
}

func ExampleNewDispatcher_updateInterval() {
	dispatcher := callcache.NewDispatcher[int](1*time.Minute, 1*time.Nanosecond)

	wg := sync.WaitGroup{}
	results := make([]any, 3)
	for i := range results {
		wg.Add(1)
		results[i], _ = dispatcher.Do("key", func() (int, error) {
			defer wg.Done()
			fmt.Printf("Do: #%d\n", i+1)
			return i + 1, nil
		})
		wg.Wait()
		time.Sleep(1 * time.Nanosecond)
	}

	for _, v := range results {
		fmt.Println(v)
	}
	// Output:
	// Do: #1
	// Do: #2
	// Do: #3
	// 1
	// 1
	// 2
}

func ExampleDispatcher_Do_multiple() {
	dispatcher := callcache.NewDispatcher[int](1*time.Minute, 10*time.Second)

	results := make([]any, 3)
	for i := range results {
		results[i], _ = dispatcher.Do("key", func() (int, error) {
			fmt.Printf("Do: #%d\n", i+1)
			return i + 1, nil
		})
	}

	for _, v := range results {
		fmt.Println(v)
	}
	// Output:
	// Do: #1
	// 1
	// 1
	// 1
}

func ExampleDispatcher_Do_concurrentSameKey() {
	dispatcher := callcache.NewDispatcher[int32](1*time.Minute, 10*time.Second)

	var value atomic.Int32

	wg := sync.WaitGroup{}
	results := make([]any, 3)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			results[i], _ = dispatcher.Do("key", func() (int32, error) {
				fmt.Println("Do")
				return value.Add(1), nil
			})
			wg.Done()
		}(i)
	}
	wg.Wait()

	for _, v := range results {
		fmt.Println(v)
	}
	// Output:
	// Do
	// 1
	// 1
	// 1
}

func ExampleDispatcher_Do_concurrentDifferentKeys() {
	dispatcher := callcache.NewDispatcher[int](1*time.Minute, 10*time.Second)

	wg := sync.WaitGroup{}
	results := make([]any, 3)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			results[i], _ = dispatcher.Do(fmt.Sprintf("key%d", i+1), func() (int, error) {
				fmt.Printf("Do: #%d\n", i+1)
				return i + 1, nil
			})
			wg.Done()
		}(i)
	}
	wg.Wait()

	for _, v := range results {
		fmt.Println(v)
	}
	// Unordered output:
	// Do: #1
	// Do: #2
	// Do: #3
	// 1
	// 2
	// 3
}

func ExampleDispatcher_Remove() {
	dispatcher := callcache.NewDispatcher[int](1*time.Minute, 10*time.Second)

	v1, _ := dispatcher.Do("key", func() (int, error) {
		fmt.Println("Do: #1")
		return 1, nil
	})
	dispatcher.Remove("key")
	v2, _ := dispatcher.Do("key", func() (int, error) {
		fmt.Println("Do: #2")
		return 2, nil
	})

	fmt.Println(v1)
	fmt.Println(v2)
	// Output:
	// Do: #1
	// Do: #2
	// 1
	// 2
}
