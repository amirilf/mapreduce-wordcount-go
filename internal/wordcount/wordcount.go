// Package wordcount implements word counting as a MapReduce job.
package wordcount

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/amirilf/mapreduce-wordcount-go/internal/mapreduce"
)

const apostrophes = "'’"

// Map emits a ("word", "1") pair for every word in contents.
func Map(contents string) []mapreduce.KeyValue {
	words := splitWords(contents)
	kvs := make([]mapreduce.KeyValue, len(words))
	for i, w := range words {
		kvs[i] = mapreduce.KeyValue{Key: w, Value: "1"}
	}
	return kvs
}

// Reduce returns the total count of a word: the sum of its values.
func Reduce(_ string, values []string) string {
	sum := 0
	for _, v := range values {
		n, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}
		sum += n
	}
	return strconv.Itoa(sum)
}

// splitWords splits text into lower-cased words made of letters.
// Apostrophes are kept inside a word (don't) and trimmed from its edges.
func splitWords(text string) []string {
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !strings.ContainsRune(apostrophes, r)
	})
	var words []string
	for _, f := range fields {
		if w := strings.Trim(f, apostrophes); w != "" {
			words = append(words, strings.ToLower(w))
		}
	}
	return words
}
