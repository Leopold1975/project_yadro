package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type TestCase struct {
	In       string
	Expected []string
}

var testcases = []TestCase{
	{
		In:       "follower brings bunch of questions",
		Expected: []string{"follow", "bring", "bunch", "question"},
	},
	{
		In:       "i'll follow you as long as you are following me",
		Expected: []string{"follow", "long"},
	},
	{
		In:       "follower, brings; !bunch of. questions?",
		Expected: []string{"follow", "bring", "bunch", "question"},
	},
	{
		In:       "he's coming right for us",
		Expected: []string{"come", "right"},
	},
	{
		In:       "we're going to your place",
		Expected: []string{"go", "place"},
	},
	{
		In:       "we are going to your place",
		Expected: []string{"go", "place"},
	},
	{
		In:       "",
		Expected: []string{},
	},
	{
		In:       "aaaaaaaa",
		Expected: []string{"aaaaaaaa"},
	},
	{
		In:       "running runners ran",
		Expected: []string{"run", "runner", "ran"},
	},
	{
		In:       "he is I am you are she's I'm we're",
		Expected: []string{},
	},
	// {
	// 	In:       "the tallest building in the city",
	// 	Expected: []string{"tall", "build", "city"},
	// },
	// {
	// 	In:       "happiness's root passed by happenstance",
	// 	Expected: []string{"happy", "root", "pass", "happenstance"},
	// },
	{
		In:       "computations compute computed computing computer's",
		Expected: []string{"comput"},
	},
	{
		In:       "origin original originated",
		Expected: []string{"origin"},
	},
	{
		In:       "as soon as possible with an passion",
		Expected: []string{"soon", "possibl", "passion"},
	},
	{
		In:       "Conor Mc'Gregor  is running away",
		Expected: []string{"conor", "mc'gregor", "run", "away"},
	},
	{
		In:       "Conor Mc'Gregor's running away",
		Expected: []string{"conor", "mc'gregor", "run", "away"},
	},
	{
		In:       "April O'Neil likes pizza",
		Expected: []string{"april", "o'neil", "like", "pizza"},
	},
	{
		In:       "Conor Mc'Gregor've done this",
		Expected: []string{"conor", "mc'gregor"},
	},
	{
		In:       "I  wouldn't've done this",
		Expected: []string{},
	},
}

func TestStemWords(t *testing.T) {
	t.Parallel()
	t.Run("test snowball", func(t *testing.T) {
		for _, tc := range testcases {
			words, err := StemWordsSnowball(tc.In)

			require.NoError(t, err)
			require.ElementsMatch(t, words, tc.Expected, "actual 	%v 	expected %v", words, tc.Expected)
		}
	})
	t.Run("test porter2", func(t *testing.T) {
		for _, tc := range testcases {
			words, err := StemWordsPorter(tc.In)

			require.NoError(t, err)
			require.ElementsMatch(t, words, tc.Expected, "actual 	%v 	expected %v", words, tc.Expected)
		}
	})
}

func BenchmarkSnowball(b *testing.B) {
	b.ResetTimer()
	b.SetBytes(1 << 5) // 32 B per op
	for i := 0; i < b.N; i++ {
		for _, tc := range testcases {
			_, _ = StemWordsSnowball(tc.In)
		}
	}
}

func BenchmarkPorter(b *testing.B) {
	b.ResetTimer()
	b.SetBytes(1 << 5)
	for i := 0; i < b.N; i++ {
		for _, tc := range testcases {
			_, _ = StemWordsPorter(tc.In)
		}
	}
}
