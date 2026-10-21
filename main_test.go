package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
	}{
		{"file names", []string{"testdata/input1.txt", "testdata/input2.txt"}},
		{"glob pattern", []string{"testdata/input*.txt"}},
	}
	want, err := os.ReadFile("testdata/output.golden")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "output.txt")
			var stderr strings.Builder
			if err := run(tt.patterns, 3, output, &stderr); err != nil {
				t.Fatal(err)
			}
			if stderr.Len() > 0 {
				t.Errorf("unexpected stderr: %s", stderr.String())
			}
			got, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Errorf("output.txt:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

func TestRunMissingFile(t *testing.T) {
	output := filepath.Join(t.TempDir(), "output.txt")
	var stderr strings.Builder

	patterns := []string{"testdata/input1.txt", "testdata/missing.txt", "testdata/input2.txt"}
	if err := run(patterns, 3, output, &stderr); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(stderr.String(), "testdata/missing.txt") {
		t.Errorf("stderr = %q, want an error naming the missing file", stderr.String())
	}
	if _, err := os.Stat(output); err != nil {
		t.Errorf("output not written: %v", err)
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
