package mapreduce

import (
	"slices"
	"testing"
)

func TestPartition(t *testing.T) {
	keys := []string{"", "a", "hello", "world", "the", "don't", "1895"}
	for _, nReduce := range []int{1, 3, 10} {
		for _, k := range keys {
			p := partition(k, nReduce)
			if p < 0 || p >= nReduce {
				t.Errorf("partition(%q, %d) = %d, out of range", k, nReduce, p)
			}
			if again := partition(k, nReduce); again != p {
				t.Errorf("partition(%q, %d) not stable: %d then %d", k, nReduce, p, again)
			}
		}
	}
}

func TestShuffle(t *testing.T) {
	const nReduce = 3
	in := make(chan KeyValue)
	go func() {
		defer close(in)
		for _, w := range []string{"hello", "fox", "hello", "dog", "hello", "fox"} {
			in <- KeyValue{Key: w, Value: "1"}
		}
	}()

	partitions := shuffle(in, nReduce)

	if len(partitions) != nReduce {
		t.Fatalf("got %d partitions, want %d", len(partitions), nReduce)
	}
	want := map[string][]string{
		"hello": {"1", "1", "1"},
		"fox":   {"1", "1"},
		"dog":   {"1"},
	}
	total := 0
	for i, p := range partitions {
		total += len(p)
		for key := range p {
			if partition(key, nReduce) != i {
				t.Errorf("key %q in partition %d, want %d", key, i, partition(key, nReduce))
			}
		}
	}
	if total != len(want) {
		t.Errorf("got %d keys in total, want %d", total, len(want))
	}
	for key, values := range want {
		if got := partitions[partition(key, nReduce)][key]; !slices.Equal(got, values) {
			t.Errorf("%q -> %q, want %q", key, got, values)
		}
	}
}
