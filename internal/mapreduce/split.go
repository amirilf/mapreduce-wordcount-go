package mapreduce

import "strings"

// split cuts text into chunks of at most size bytes. Each chunk ends at a line
// break, so no line is cut; a single line longer than size stays one chunk.
// A size of 0 or less returns the whole text as one chunk.
func split(text string, size int) []string {
	if size <= 0 || len(text) <= size {
		return []string{text}
	}
	var chunks []string
	for len(text) > size {
		cut := strings.LastIndexByte(text[:size], '\n') + 1
		if cut == 0 {
			cut = strings.IndexByte(text, '\n') + 1
			if cut == 0 {
				cut = len(text)
			}
		}
		chunks = append(chunks, text[:cut])
		text = text[cut:]
	}
	if text != "" {
		chunks = append(chunks, text)
	}
	return chunks
}
