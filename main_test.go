package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	output := filepath.Join(t.TempDir(), "output.txt")
	files := []string{"testdata/input1.txt", "testdata/missing.txt", "testdata/input2.txt"}
	var stderr strings.Builder

	if err := run(files, 3, output, &stderr); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(stderr.String(), "testdata/missing.txt") {
		t.Errorf("stderr = %q, want an error naming the missing file", stderr.String())
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/output.golden")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("output.txt:\n%s\nwant:\n%s", got, want)
	}
}

func TestRunNoReadableFile(t *testing.T) {
	output := filepath.Join(t.TempDir(), "output.txt")
	var stderr strings.Builder

	if err := run([]string{"testdata/missing.txt"}, 3, output, &stderr); err == nil {
		t.Error("want an error when no file can be read")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Errorf("%s was written, want no output file", output)
	}
}
