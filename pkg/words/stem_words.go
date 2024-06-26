package words

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kljensen/snowball"
	"github.com/surgebase/porter2"
	"golang.org/x/text/unicode/norm"
)

const (
	EngLang = "english"
)

func StemWords(phrase string) ([]string, error) {
	return StemWordsPorter(phrase)
}

// StemWordsPorter is used for turning a phrase into a list of stemmed words
// and uses porter lib for it. Example:
//
// var s string
//
// flag.StringVar(&s, "s", "", "a sentence to normalize")
//
// stemmed, err := StemWordsPorter(s)
//
// fmt.Println(stemmed).
func StemWordsPorter(phrase string) ([]string, error) {
	words := GetWords(phrase)
	uniqueWords := make(map[string]struct{}, len(words))

	for _, w := range words {
		w = porter2.Stem(w)

		uniqueWords[w] = struct{}{}
	}

	wordsResult := make([]string, 0, len(uniqueWords))
	for w := range uniqueWords {
		wordsResult = append(wordsResult, w)
	}

	return wordsResult, nil
}

func StemWordsSnowball(phrase string) ([]string, error) {
	words := GetWords(phrase)
	uniqueWords := make(map[string]struct{}, len(words))

	for _, w := range words {
		w, err := snowball.Stem(w, EngLang, true)
		if err != nil {
			return nil, fmt.Errorf("snowball stemming error: %w", err)
		}

		uniqueWords[w] = struct{}{}
	}

	wordsResult := make([]string, 0, len(uniqueWords))
	for w := range uniqueWords {
		wordsResult = append(wordsResult, w)
	}

	return wordsResult, nil
}

// GetWords function that gives back slice of words
// except prepositions, pronouns and punctuation marks.
func GetWords(phrase string) []string {
	// See bench_larger_regexp and bench_smaller_regexp to finc out the difference in
	// performance of these two approaches
	// re := regexp.MustCompile(`([\,\;\.\?\!\:\&]+|n't|'ve|'re|'m|'ll|'d|'s)`)
	re := regexp.MustCompile(`[\,\;\.\?\!\:\&\[\]\{\}\"\*\(\)]+`)
	cleanedString := re.ReplaceAllString(phrase, "")

	words := strings.Fields(cleanedString)
	words = splitCompositeWords(words)

	wordsResult := make([]string, 0, len(words))

	for _, word := range words {
		word = string(norm.NFKD.Bytes([]byte(word)))
		word = strings.ToLower(word)

		if i := strings.Index(word, "'"); i != -1 {
			// Part of "larger regexp" solution
			// if len(word[i:]) < 1 {
			// 	word = word[:i]
			// }
			lenOfShorts := 3 // such as 'll, 've, 't, etc.

			switch {
			case len(word[i:]) <= lenOfShorts:
				word = word[:i]
			case strings.Contains(word[i+1:], "'"):
				j := strings.Index(word[i+1:], "'")
				word = word[:j+i+1] // equivalent to word = word[:i+1] + word[i+1:][:j]
			}

			if strings.Contains(word, "'") && len(word) > 3 && word[len(word)-3:] == "n't" {
				word = word[:len(word)-3]
			}
		}

		if _, ok := stopWords[word]; ok {
			continue
		}

		wordsResult = append(wordsResult, word)
	}

	return wordsResult
}

func splitCompositeWords(words []string) []string {
	for i, w := range words {
		wrds := strings.Split(w, "/")
		if len(wrds) != 0 {
			words[i] = wrds[0]
			words = append(words, wrds[1:]...)
		}
	}

	return words
}
