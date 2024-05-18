package usecase_test

import (
	"testing"

	"github.com/Leopold1975/yadro_app/internal/usecase"
	"github.com/stretchr/testify/require"
)

func TestGetTopIDs(t *testing.T) {
	tests := []struct {
		name     string
		results  [][]string
		expected []string
	}{
		{
			name:     "same strings",
			results:  [][]string{{"1", "2"}, {"2", "1"}, {"1", "3"}},
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "different strings",
			results:  [][]string{{"1"}, {"2"}, {"3"}},
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "with unique and repeated",
			results:  [][]string{{"1", "2", "2", "3"}, {"2"}, {"3", "3", "1"}},
			expected: []string{"2", "3", "1"},
		},
		{
			name: "1 to 10 are most frequest",
			results: [][]string{
				{"1", "11", "12", "13", "14", "15", "16", "17", "18", "19"},
				{"1", "2", "21", "22", "23", "24", "25", "26", "27", "28"},
				{"1", "2", "3", "31", "32", "33", "34", "35", "36", "37"},
				{"1", "2", "3", "4", "41", "42", "43", "44", "45", "46"},
				{"1", "2", "3", "4", "5", "51", "52", "53", "54", "55"},
				{"1", "2", "3", "4", "5", "6", "61", "62", "63", "64"},
				{"1", "2", "3", "4", "5", "6", "7", "71", "72", "73"},
				{"1", "2", "3", "4", "5", "6", "7", "8", "81", "82"},
				{"1", "2", "3", "4", "5", "6", "7", "8", "9", "91"},
				{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "10"},
			},
			expected: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := usecase.GetTopIDs(tc.results...)
			require.ElementsMatch(t, got, tc.expected, "getTopIDs() = %v, want %v", got, tc.expected)
		})
	}
}
