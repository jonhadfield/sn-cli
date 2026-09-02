package main

import (
	"strings"
	"testing"

	"github.com/jonhadfield/gosn-v2/items"
	"github.com/stretchr/testify/assert"
)

func TestSearchNotes(t *testing.T) {
	// Create mock notes
	note1, _ := items.NewNote("Test Note One", "This is a test note with some content", nil)
	// Lower-case "test" so the content-search case finds this note while the
	// title-only and case-sensitive ("Test") cases still do not.
	note2, _ := items.NewNote("Another Note", "Different content here, with a test reference", nil)
	note3, _ := items.NewNote("Test Note Two", "More test content", nil)

	notes := items.Items{&note1, &note2, &note3}

	tests := []struct {
		name          string
		query         string
		searchContent bool
		fuzzy         bool
		caseSensitive bool
		expectedCount int
	}{
		{
			name:          "search title only",
			query:         "test",
			searchContent: false,
			fuzzy:         false,
			caseSensitive: false,
			expectedCount: 2, // note1 and note3 have "test" in title
		},
		{
			name:          "search content and title",
			query:         "test",
			searchContent: true,
			fuzzy:         false,
			caseSensitive: false,
			expectedCount: 3, // all notes have "test" somewhere
		},
		{
			name:          "case sensitive search",
			query:         "Test",
			searchContent: true,
			fuzzy:         false,
			caseSensitive: true,
			expectedCount: 2, // note1 and note3 have "Test" with capital T
		},
		{
			name:          "no matches",
			query:         "xyz123",
			searchContent: true,
			fuzzy:         false,
			caseSensitive: false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := searchNotes(notes, tt.query, tt.searchContent, tt.fuzzy, tt.caseSensitive)
			assert.Equal(t, tt.expectedCount, len(results), "Unexpected number of results")
		})
	}
}

func TestGenerateSearchPreview(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		query         string
		caseSensitive bool
		expectMatch   bool
		// Text shorter than the context window is returned whole, so only
		// long text is windowed down and elided.
		expectTruncated bool
	}{
		{
			name:          "match found",
			text:          "The quick brown fox jumps over the lazy dog",
			query:         "fox",
			caseSensitive: false,
			expectMatch:   true,
		},
		{
			name:          "no match",
			text:          "The quick brown dog",
			query:         "cat",
			caseSensitive: false,
		},
		{
			name:          "case sensitive match",
			text:          "The Quick Brown Fox",
			query:         "Quick",
			caseSensitive: true,
			expectMatch:   true,
		},
		{
			name:          "case sensitive no match",
			text:          "The Quick Brown Fox",
			query:         "quick",
			caseSensitive: true,
		},
		{
			name:            "long text windowed around match",
			text:            strings.Repeat("filler words ", 20) + "needle" + strings.Repeat(" trailing words", 20),
			query:           "needle",
			caseSensitive:   false,
			expectMatch:     true,
			expectTruncated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preview := generateSearchPreview(tt.text, tt.query, tt.caseSensitive)
			assert.NotEmpty(t, preview, "Preview should not be empty")

			if tt.expectMatch {
				// The preview is centred on the match, so it must contain it.
				haystack, needle := preview, tt.query
				if !tt.caseSensitive {
					haystack, needle = strings.ToLower(haystack), strings.ToLower(needle)
				}

				assert.Contains(t, haystack, needle, "Preview should contain the match")
			}

			if tt.expectTruncated {
				assert.NotEqual(t, tt.text, preview, "Long text should be windowed")
				assert.Contains(t, preview, "...", "Windowed preview should be elided")
			}
		})
	}
}

func TestSearchResult_Sorting(t *testing.T) {
	note1, _ := items.NewNote("Low Score", "content", nil)
	note2, _ := items.NewNote("Test Title", "test content", nil)
	note3, _ := items.NewNote("Another", "random", nil)

	notes := items.Items{&note1, &note2, &note3}

	results := searchNotes(notes, "test", true, false, false)

	// Results should be sorted by score (title matches score higher)
	assert.True(t, len(results) > 0, "Should have results")
	if len(results) >= 2 {
		assert.Greater(t, results[0].Score, results[1].Score,
			"First result should have higher score than second")
	}
}

func TestSearchResult_MatchTypes(t *testing.T) {
	titleMatch, _ := items.NewNote("Search Test Title", "Different content", nil)
	textMatch, _ := items.NewNote("Other Title", "This has search test in the text", nil)
	bothMatch, _ := items.NewNote("Search Test", "Also has search test here", nil)

	notes := items.Items{&titleMatch, &textMatch, &bothMatch}

	results := searchNotes(notes, "search test", true, false, false)

	assert.Equal(t, 3, len(results), "Should match all three notes")

	// Find the result that matches both title and text
	var bothMatchResult *SearchResult
	for i := range results {
		if results[i].MatchInTitle && results[i].MatchInBody {
			bothMatchResult = &results[i]
			break
		}
	}

	assert.NotNil(t, bothMatchResult, "Should have a result matching both title and body")
	assert.True(t, bothMatchResult.MatchInTitle)
	assert.True(t, bothMatchResult.MatchInBody)
}

func TestTruncateTitle(t *testing.T) {
	tests := []struct {
		name   string
		title  string
		maxLen int
	}{
		{
			name:   "short title",
			title:  "Short",
			maxLen: 10,
		},
		{
			name:   "long title",
			title:  "This is a very long title that needs to be truncated",
			maxLen: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateTitle(tt.title, tt.maxLen)
			assert.LessOrEqual(t, len(result), tt.maxLen+3, "Truncated title too long")
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		maxLen int
	}{
		{
			name:   "short string",
			str:    "Short",
			maxLen: 10,
		},
		{
			name:   "long string",
			str:    "This is a very long string that definitely needs truncation",
			maxLen: 20,
		},
		{
			name:   "exact length",
			str:    "Exactly20Characters!",
			maxLen: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.str, tt.maxLen)
			assert.LessOrEqual(t, len(result), tt.maxLen+3, "Truncated string too long")
		})
	}
}
