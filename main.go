// Command wordcount counts how often each word occurs across text files using
// MapReduce, and writes the counts sorted by word to output.txt.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/amirilf/mapreduce-wordcount-go/internal/mapreduce"
	"github.com/amirilf/mapreduce-wordcount-go/internal/wordcount"
)

var defaultFiles = []string{"input1.txt", "input2.txt", "input3.txt", "input4.txt", "input5.txt"}

func main() {
	nReduce := flag.Int("nreduce", 4, "number of reduce partitions")
	flag.Parse()
	if *nReduce < 1 {
		fmt.Fprintln(os.Stderr, "-nreduce must be at least 1")
		os.Exit(2)
	}

	files := flag.Args()
	if len(files) == 0 {
		files = defaultFiles
	}
	if err := run(files, *nReduce, "output.txt", os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run counts the words of files into output. Files that cannot be read are
// reported to stderr and skipped; if nothing is left to count, output is not written.
func run(files []string, nReduce int, output string, stderr io.Writer) error {
	results, err := mapreduce.Run(files, wordcount.Map, wordcount.Reduce, nReduce)
	if err != nil {
		fmt.Fprintln(stderr, err)
		if len(results) == 0 {
			return fmt.Errorf("no words counted, %s not written", output)
		}
	}
	var b strings.Builder
	for _, kv := range results {
		fmt.Fprintf(&b, "%s: %s\n", kv.Key, kv.Value)
	}
	return os.WriteFile(output, []byte(b.String()), 0o644)
}
