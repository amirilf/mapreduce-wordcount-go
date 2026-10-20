package wordcount

import (
	"slices"
	"testing"

	"github.com/amirilf/mapreduce-wordcount-go/internal/mapreduce"
)

func TestWords(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{"empty", "", nil},
		{"only separators", " \n\t.,;!? -- ", nil},
		{"case insensitive", "The THE the", []string{"the", "the", "the"}},
		{"punctuation stripped", "Hello, world! (yes)... \"quoted\"", []string{"hello", "world", "yes", "quoted"}},
		{"newlines and tabs", "one\ntwo\tthree\r\nfour", []string{"one", "two", "three", "four"}},
		{"contractions kept", "Don't, I'll couldn't", []string{"don't", "i'll", "couldn't"}},
		{"quotes stripped", "'tis 'hello' ’curly' '", []string{"tis", "hello", "curly"}},
		{"hyphen separates", "re-use Gutenberg-tm", []string{"re", "use", "gutenberg", "tm"}},
		{"numbers are not words", "chapter 12, 1895.", []string{"chapter"}},
		{"contractions with i", "I'd I'm I've", []string{"i'd", "i'm", "i've"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Words(tt.text); !slices.Equal(got, tt.want) {
				t.Errorf("Words(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

func TestReduce(t *testing.T) {
	tests := []struct {
		values []string
		want   string
	}{
		{[]string{"1"}, "1"},
		{[]string{"1", "1", "1"}, "3"},
		{[]string{"2", "40"}, "42"},
	}
	for _, tt := range tests {
		if got := Reduce("word", tt.values); got != tt.want {
			t.Errorf("Reduce(%q) = %s, want %s", tt.values, got, tt.want)
		}
	}
}

func TestMap(t *testing.T) {
	got := Map("the quick brown fox jumps over the lazy dog")
	words := []string{"the", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog"}
	if len(got) != len(words) {
		t.Fatalf("Map emitted %d pairs, want %d", len(got), len(words))
	}
	for i, w := range words {
		if want := (mapreduce.KeyValue{Key: w, Value: "1"}); got[i] != want {
			t.Errorf("pair %d = %v, want %v", i, got[i], want)
		}
	}
}
