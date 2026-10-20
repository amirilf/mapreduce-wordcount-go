// Package mapreduce runs a MapReduce job on a single machine using goroutines and channels.
package mapreduce

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
)

// KeyValue is an intermediate pair emitted by a MapFunc, or a final result.
type KeyValue struct {
	Key   string
	Value string
}

// MapFunc turns the contents of one input into key/value pairs.
type MapFunc func(contents string) []KeyValue

// ReduceFunc combines all values of one key into its final value.
type ReduceFunc func(key string, values []string) string

// Run maps every file in its own goroutine, shuffles the pairs into nReduce
// partitions and reduces each partition in its own goroutine. It returns the
// results sorted by key. Files that cannot be read are skipped and reported in the returned error.
func Run(files []string, mapf MapFunc, reducef ReduceFunc, nReduce int) ([]KeyValue, error) {
	if nReduce < 1 {
		return nil, fmt.Errorf("nReduce must be at least 1, got %d", nReduce)
	}

	intermediate := make(chan KeyValue)
	readErrs := make(chan error, len(files))
	var mappers sync.WaitGroup
	for _, file := range files {
		mappers.Add(1)
		go func() {
			defer mappers.Done()
			contents, err := os.ReadFile(file)
			if err != nil {
				readErrs <- err
				return
			}
			for _, kv := range mapf(string(contents)) {
				intermediate <- kv
			}
		}()
	}
	go func() {
		mappers.Wait()
		close(intermediate)
		close(readErrs)
	}()

	partitions := shuffle(intermediate, nReduce)
	var errs []error
	for err := range readErrs {
		errs = append(errs, err)
	}

	results := make(chan KeyValue)
	var reducers sync.WaitGroup
	for _, p := range partitions {
		reducers.Add(1)
		go func() {
			defer reducers.Done()
			for key, values := range p {
				results <- KeyValue{Key: key, Value: reducef(key, values)}
			}
		}()
	}
	go func() {
		reducers.Wait()
		close(results)
	}()

	var out []KeyValue
	for kv := range results {
		out = append(out, kv)
	}
	slices.SortFunc(out, func(a, b KeyValue) int { return strings.Compare(a.Key, b.Key) })
	return out, errors.Join(errs...)
}
