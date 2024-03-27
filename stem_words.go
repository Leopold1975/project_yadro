package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/kljensen/snowball"
	"github.com/surgebase/porter2"
)

func main() {
	s := flag.String("s", "", "a sentence to normalize")
	flag.Parse()

	stemmed, err := StemWordsPorter(*s)
	if err != nil {
		log.Fatalf("stem words error: %s", err.Error())
	}

	fmt.Println(stemmed) //nolint:forbidigo
}

const (
	EngLang = "english"
)

func StemWordsPorter(phrase string) ([]string, error) { //nolint:unparam
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
	re := regexp.MustCompile(`[\,\;\.\?\!\:\&]+`)
	cleanedString := re.ReplaceAllString(phrase, "")

	words := strings.Fields(cleanedString)
	wordsResult := make([]string, 0, len(words))

	for _, word := range words {
		word = strings.ToLower(word)
		if i := strings.Index(word, "'"); i != -1 {
			word = word[:i]
		}

		if _, ok := stopWords[word]; ok {
			continue
		}

		wordsResult = append(wordsResult, word)
	}

	return wordsResult
}
