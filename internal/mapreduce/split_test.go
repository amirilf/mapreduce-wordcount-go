package mapreduce

import (
	"slices"
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name string
		text string
		size int
		want []string
	}{
		{"no chunking", "a\nb\nc\n", 0, []string{"a\nb\nc\n"}},
		{"fits in one chunk", "a\nb\n", 10, []string{"a\nb\n"}},
		{"cut at line breaks", "aa\nbb\ncc\n", 6, []string{"aa\nbb\n", "cc\n"}},
		{"one line per chunk", "aa\nbb\ncc", 3, []string{"aa\n", "bb\n", "cc"}},
		{"long line stays whole", "aaaaaa\nb\n", 3, []string{"aaaaaa\n", "b\n"}},
		{"no line break at all", "aaaaaa", 3, []string{"aaaaaa"}},
		{"empty", "", 3, []string{""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := split(tt.text, tt.size); !slices.Equal(got, tt.want) {
				t.Errorf("split(%q, %d) = %q, want %q", tt.text, tt.size, got, tt.want)
			}
		})
	}
}

func TestSplitKeepsText(t *testing.T) {
	text := strings.Repeat("the quick brown fox\njumps over\n\nthe lazy dog\n", 50)
	for size := 1; size <= len(text)+1; size += 7 {
		chunks := split(text, size)
		if got := strings.Join(chunks, ""); got != text {
			t.Fatalf("size %d: chunks do not add up to the text", size)
		}
		for _, c := range chunks {
			if len(c) > size && strings.Count(c, "\n") > 1 {
				t.Fatalf("size %d: chunk of %d bytes holds more than one line", size, len(c))
			}
		}
	}
}
