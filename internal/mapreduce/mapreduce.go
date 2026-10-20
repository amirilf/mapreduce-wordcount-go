// Package mapreduce runs a MapReduce job on a single machine using goroutines and channels.
package mapreduce

// KeyValue is an intermediate pair emitted by a MapFunc.
type KeyValue struct {
	Key   string
	Value string
}

// MapFunc turns the contents of one input into key/value pairs.
type MapFunc func(contents string) []KeyValue
