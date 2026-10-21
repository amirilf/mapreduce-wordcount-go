// Command wordcount counts how often each word occurs across text files using
// MapReduce, and writes the counts sorted by word to output.txt.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/amirilf/mapreduce-wordcount-go/internal/mapreduce"
	"github.com/amirilf/mapreduce-wordcount-go/internal/wordcount"
)

var defaultFiles = []string{"input1.txt", "input2.txt", "input3.txt", "input4.txt", "input5.txt"}

func main() {
	nReduce := flag.Int("nreduce", 4, "number of reduce partitions, one reduce goroutine each")
	chunkSize := flag.Int("chunk", 0, "max bytes per map goroutine; 0 means one map goroutine per file")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: go run main.go [-nreduce n] [-chunk bytes] [file-or-pattern ...]")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *nReduce < 1 || *chunkSize < 0 {
		flag.Usage()
		os.Exit(2)
	}

	patterns := flag.Args()
	if len(patterns) == 0 {
		patterns = defaultFiles
	}
	if err := run(patterns, *nReduce, *chunkSize, "output.txt", os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run counts the words of all files matching patterns into output. Patterns
// that match nothing and files that cannot be read are reported to stderr and
// skipped; if no words are left, output is not written.
func run(patterns []string, nReduce, chunkSize int, output string, stderr io.Writer) error {
	var files []string
	for _, p := range patterns {
		matches, err := filepath.Glob(p)
		if err != nil {
			fmt.Fprintf(stderr, "%s: %v\n", p, err)
			continue
		}
		if len(matches) == 0 {
			fmt.Fprintf(stderr, "%s: no such file\n", p)
			continue
		}
		files = append(files, matches...)
	}

	results, err := mapreduce.Run(files, wordcount.Map, wordcount.Reduce, nReduce, chunkSize)
	if err != nil {
		fmt.Fprintln(stderr, err)
	}
	if len(results) == 0 {
		return fmt.Errorf("no words counted, %s not written", output)
	}

	var b strings.Builder
	for _, kv := range results {
		fmt.Fprintf(&b, "%s: %s\n", kv.Key, kv.Value)
	}
	return os.WriteFile(output, []byte(b.String()), 0o644)
}
