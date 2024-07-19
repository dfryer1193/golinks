package search

import (
	"slices"
)

type SearchResult struct {
	Value string
	Score int
}

func StringSearch(query string, options []string) []SearchResult {
	results := []SearchResult{}
	for _, val := range options {
		score := computeLevenshtein(query, val)
		result := SearchResult{
			Value: val,
			Score: score,
		}
		results = append(results, result)
	}

	slices.SortStableFunc(results, func(a, b SearchResult) int {
		return -1 * (a.Score - b.Score) // Lower scores have closer matches
	})

	return results
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
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}
	return min
}
