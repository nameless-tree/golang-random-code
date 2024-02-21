package searcher

import (
	"regexp"
)

func removePunctuation(word string) string {
	re := regexp.MustCompile(`[^\p{L}\p{N}]`)
	return re.ReplaceAllString(word, "")
}

func addWordToMap(words map[string]map[int]struct{}, word string, value int) {
	// If a map for a given word does not yet exist, create it
	if words[word] == nil {
		words[word] = make(map[int]struct{})
	}

	words[word][value] = struct{}{}
}
