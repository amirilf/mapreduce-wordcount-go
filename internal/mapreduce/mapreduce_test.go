package mapreduce

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func splitWords(contents string) []KeyValue {
	var kvs []KeyValue
	for _, w := range strings.Fields(contents) {
		kvs = append(kvs, KeyValue{Key: w, Value: "1"})
	}
	return kvs
}

func countValues(_ string, values []string) string {
	return strconv.Itoa(len(values))
}

func writeFiles(t *testing.T, contents ...string) []string {
	t.Helper()
	dir := t.TempDir()
	var files []string
	for i, c := range contents {
		name := filepath.Join(dir, "input"+strconv.Itoa(i+1)+".txt")
		if err := os.WriteFile(name, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
		files = append(files, name)
	}
	return files
}

func TestRun(t *testing.T) {
	files := writeFiles(t,
		"the quick brown fox\njumps over\nthe lazy dog\n",
		"the dog barks\nand the fox\nruns away\n",
	)
	want := []KeyValue{
		{"and", "1"}, {"away", "1"}, {"barks", "1"}, {"brown", "1"}, {"dog", "2"}, {"fox", "2"},
		{"jumps", "1"}, {"lazy", "1"}, {"over", "1"}, {"quick", "1"}, {"runs", "1"}, {"the", "4"},
	}
	for _, nReduce := range []int{1, 3, 8} {
		for _, chunkSize := range []int{0, 1, 12, 1000} {
			got, err := Run(files, splitWords, countValues, nReduce, chunkSize)
			if err != nil {
				t.Fatalf("nReduce=%d chunkSize=%d: %v", nReduce, chunkSize, err)
			}
			if !slices.Equal(got, want) {
				t.Errorf("nReduce=%d chunkSize=%d:\n got %v\nwant %v", nReduce, chunkSize, got, want)
			}
		}
	}
}

func TestRunMissingFile(t *testing.T) {
	files := writeFiles(t, "a b", "b c")
	missing := filepath.Join(t.TempDir(), "missing.txt")
	files = []string{files[0], missing, files[1]}

	got, err := Run(files, splitWords, countValues, 2, 0)

	if !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), missing) {
		t.Errorf("err = %v, want a not-exist error naming %s", err, missing)
	}
	want := []KeyValue{{"a", "1"}, {"b", "2"}, {"c", "1"}}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRunInvalidNReduce(t *testing.T) {
	if _, err := Run(nil, splitWords, countValues, 0, 0); err == nil {
		t.Error("nReduce=0: want an error")
	}
}
