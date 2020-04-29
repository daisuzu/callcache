package callcache_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/daisuzu/callcache"
)

func ExampleDispatcher() {
	dispatcher := callcache.NewDispatcher(1*time.Minute, 10*time.Second)

	v, err := dispatcher.Do("key", func() (interface{}, error) {
		return "example", nil
	})

	fmt.Println(v, err)
	// Output:
	// example <nil>
}

func ExampleNewDispatcher_expiration() {
	dispatcher := callcache.NewDispatcher(1*time.Nanosecond, 0)

	wg := sync.WaitGroup{}
	results := make([]interface{}, 3)
	for i := range results {
		wg.Add(1)
		results[i], _ = dispatcher.Do("key", func() (interface{}, error) {
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
	dispatcher := callcache.NewDispatcher(1*time.Minute, 1*time.Nanosecond)

	wg := sync.WaitGroup{}
	results := make([]interface{}, 3)
	for i := range results {
		wg.Add(1)
		results[i], _ = dispatcher.Do("key", func() (interface{}, error) {
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
	dispatcher := callcache.NewDispatcher(1*time.Minute, 10*time.Second)

	results := make([]interface{}, 3)
	for i := range results {
		results[i], _ = dispatcher.Do("key", func() (interface{}, error) {
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
	dispatcher := callcache.NewDispatcher(1*time.Minute, 10*time.Second)

	var value int32

	wg := sync.WaitGroup{}
	results := make([]interface{}, 3)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			results[i], _ = dispatcher.Do("key", func() (interface{}, error) {
				fmt.Println("Do")
				return atomic.AddInt32(&value, 1), nil
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
	dispatcher := callcache.NewDispatcher(1*time.Minute, 10*time.Second)

	wg := sync.WaitGroup{}
	results := make([]interface{}, 3)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			results[i], _ = dispatcher.Do(fmt.Sprintf("key%d", i+1), func() (interface{}, error) {
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
	dispatcher := callcache.NewDispatcher(1*time.Minute, 10*time.Second)

	v1, _ := dispatcher.Do("key", func() (interface{}, error) {
		fmt.Println("Do: #1")
		return 1, nil
	})
	dispatcher.Remove("key")
	v2, _ := dispatcher.Do("key", func() (interface{}, error) {
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
