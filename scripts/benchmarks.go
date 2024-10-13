package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/umgbhalla/gokv/pkg/client"
)

const (
	numOperations = 200
	baseURL       = "http://localhost:8080"
)

func main() {
	gokv := client.New(baseURL)

	fmt.Println("Running GoKV Benchmarks...")

	// Benchmark Set operations
	setDuration := benchmarkSet(gokv)
	fmt.Printf("Set Benchmark: %d operations in %v (%.2f ops/sec)\n", numOperations, setDuration, float64(numOperations)/setDuration.Seconds())

	// Benchmark Get operations
	getDuration := benchmarkGet(gokv)
	fmt.Printf("Get Benchmark: %d operations in %v (%.2f ops/sec)\n", numOperations, getDuration, float64(numOperations)/getDuration.Seconds())

	// Benchmark Delete operations
	deleteDuration := benchmarkDelete(gokv)
	fmt.Printf("Delete Benchmark: %d operations in %v (%.2f ops/sec)\n", numOperations, deleteDuration, float64(numOperations)/deleteDuration.Seconds())

	// Benchmark concurrent operations
	concurrentDuration := benchmarkConcurrent(gokv)
	fmt.Printf("Concurrent Benchmark: %d operations in %v (%.2f ops/sec)\n", numOperations*3, concurrentDuration, float64(numOperations*3)/concurrentDuration.Seconds())
}

func benchmarkSet(gokv *client.Client) time.Duration {
	start := time.Now()
	for i := 0; i < numOperations; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		err := gokv.Set(key, value, time.Second*30)
		if err != nil {
			fmt.Printf("Error setting key %s: %v\n", key, err)
		}
	}
	return time.Since(start)
}

func benchmarkGet(gokv *client.Client) time.Duration {
	start := time.Now()
	for i := 0; i < numOperations; i++ {
		key := fmt.Sprintf("key%d", i)
		_, err := gokv.Get(key)
		if err != nil {
			fmt.Printf("Error getting key %s: %v\n", key, err)
		} 
		
	}
	return time.Since(start)
}

func benchmarkDelete(gokv *client.Client) time.Duration {
	start := time.Now()
	for i := 0; i < numOperations; i++ {
		key := fmt.Sprintf("key%d", i)
		err := gokv.Delete(key)
		if err != nil {
			fmt.Printf("Error deleting key %s: %v\n", key, err)
		}
	}
	return time.Since(start)
}

func benchmarkConcurrent(gokv *client.Client) time.Duration {
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numOperations; i++ {
		wg.Add(3)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent_key%d", i)
			value := fmt.Sprintf("concurrent_value%d", i)
			err := gokv.Set(key, value, time.Second*20)
			if err != nil {
				fmt.Printf("Error setting key %s: %v\n", key, err)
			}
		}(i)

		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent_key%d", i)
			_, err := gokv.Get(key)
			if err != nil {
				fmt.Printf("Error getting key %s: %v\n", key, err)
			}
		}(i)

		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent_key%d", rand.Intn(i+1))
			err := gokv.Delete(key)
			if err != nil {
				fmt.Printf("Error deleting key %s: %v\n", key, err)
			}
		}(i)
	}

	wg.Wait()
	return time.Since(start)
}
