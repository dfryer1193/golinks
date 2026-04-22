package search

import (
	"slices"
	"strings"
)

const shortcutEditThreshold = 3

type Result struct {
	Value string
	// Score represents the match quality:
	// - 0 for substring matches (best)
	// - Levenshtein edit distance for other matches (lower is better)
	Score int
}

// StringSearch performs fuzzy search with both Levenshtein distance and substring matching.
//
// Scoring behavior:
//   - Substring matches (case-insensitive): Score = 0 (highest priority)
//   - Levenshtein distance: Score = edit distance between query and option
//
// Only results with Score <= shortcutEditThreshold (3) are returned.
// Results are sorted by Score (lower is better).
//
// Examples:
//   - "stack" -> "stackoverflow": Score = 0 (substring match)
//   - "githb" -> "github": Score = 1 (1 edit: insert 'u')
//   - "doc" -> "docs": Score = 1 (1 edit: insert 's')
func StringSearch(query string, options []string) []Result {
	results := make([]Result, 0, len(options))
	
	for _, val := range options {
		score := computeLevenshtein(query, val)
		
		// Also check for substring matches (case-insensitive)
		// Substring matches get a score of 0 for better ranking
		if strings.Contains(strings.ToLower(val), strings.ToLower(query)) {
			score = 0
		}
		
		result := Result{
			Value: val,
			Score: score,
		}
		results = append(results, result)
	}

	slices.SortStableFunc(results, func(a, b Result) int {
		return a.Score - b.Score // Lower scores are closer matches
	})

	// Filter out results with score > threshold
	filtered := make([]Result, 0)
	for _, result := range results {
		if result.Score <= shortcutEditThreshold {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

func computeLevenshtein(query, value string) int {
	queryLen := len(query)
	valueLen := len(value)
	distanceMatrix := make([][]int, queryLen)
	for i := 0; i < queryLen; i++ {
		distanceMatrix[i] = make([]int, valueLen)
	}

	// Match empty string by dropping all characters
	for i := 0; i < queryLen; i++ {
		distanceMatrix[i][0] = i
	}

	for j := 0; j < valueLen; j++ {
		distanceMatrix[0][j] = j
	}

	for j := 1; j < valueLen; j++ {
		for i := 1; i < queryLen; i++ {
			cost := 0
			if query[i] != value[j] {
				cost = 1
			}

			distanceMatrix[i][j] = minOf(
				distanceMatrix[i-1][j]+1,
				distanceMatrix[i][j-1]+1,
				distanceMatrix[i-1][j-1]+cost,
			)
		}
	}

	return distanceMatrix[queryLen-1][valueLen-1]
}

func minOf(vars ...int) int {
	m := vars[0]

	for _, i := range vars {
		if m > i {
			m = i
		}
	}
	return m
}
