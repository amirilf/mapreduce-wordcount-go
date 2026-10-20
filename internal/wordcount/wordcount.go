// Package wordcount implements word counting as a MapReduce job.
package wordcount

import (
	"strings"
	"unicode"

	"github.com/amirilf/mapreduce-wordcount-go/internal/mapreduce"
)

const apostrophes = "'’"

// Map emits a ("word", "1") pair for every word in contents.
func Map(contents string) []mapreduce.KeyValue {
	words := Words(contents)
	kvs := make([]mapreduce.KeyValue, len(words))
	for i, w := range words {
		kvs[i] = mapreduce.KeyValue{Key: w, Value: "1"}
	}
	return kvs
}

// Words splits text into lower-cased words of letters and digits.
// Apostrophes are kept inside a word (don't) and trimmed from its edges.
func Words(text string) []string {
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune(apostrophes, r)
	})
	words := fields[:0]
	for _, f := range fields {
		if w := strings.Trim(f, apostrophes); w != "" {
			words = append(words, strings.ToLower(w))
		}
	}
	return words
}
