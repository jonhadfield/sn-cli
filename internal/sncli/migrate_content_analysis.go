package sncli

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
)

// ContentTheme represents a discovered theme from content analysis.
type ContentTheme struct {
	Name       string
	Keywords   []string
	Phrases    []string
	NoteCount  int
	Relevance  float64
	RelatedNotes []string // Note UUIDs
}

// ContentAnalyzer performs content analysis on notes to discover themes.
type ContentAnalyzer struct {
	notes           items.Items
	stopWords       map[string]bool
	minKeywordFreq  int
	minPhraseFreq   int
	minThemeNotes   int
}

// NewContentAnalyzer creates a content analyzer.
func NewContentAnalyzer(notes items.Items) *ContentAnalyzer {
	return &ContentAnalyzer{
		notes:          notes,
		stopWords:      buildStopWords(),
		minKeywordFreq: 2, // Keyword must appear in at least 2 notes
		minPhraseFreq:  2, // Phrase must appear in at least 2 notes
		minThemeNotes:  2, // Theme must have at least 2 notes (lower than tag threshold)
	}
}

// AnalyzeContent performs comprehensive content analysis.
func (ca *ContentAnalyzer) AnalyzeContent() []ContentTheme {
	// Extract all text from notes
	noteTexts := ca.extractNoteTexts()

	// Build document frequency maps
	keywordDF := ca.calculateKeywordDF(noteTexts)
	phraseDF := ca.calculatePhraseDF(noteTexts)

	// Calculate TF-IDF scores
	keywordScores := ca.calculateTFIDF(noteTexts, keywordDF)
	phraseScores := ca.calculatePhraseTFIDF(noteTexts, phraseDF)

	// Identify themes from top keywords and phrases
	themes := ca.identifyThemes(keywordScores, phraseScores, noteTexts)

	return themes
}

// extractNoteTexts extracts title and content text from all notes.
func (ca *ContentAnalyzer) extractNoteTexts() map[string]string {
	texts := make(map[string]string)

	for _, item := range ca.notes {
		if item.GetContentType() != common.SNItemTypeNote {
			continue
		}

		note, ok := item.(*items.Note)
		if !ok {
			continue
		}

		// Combine title and text with extra weight on title
		title := note.Content.GetTitle()
		text := note.Content.GetText()

		// Title words appear 3x for emphasis
		combined := title + " " + title + " " + title + " " + text
		texts[note.UUID] = strings.ToLower(combined)
	}

	return texts
}

// calculateKeywordDF calculates document frequency for keywords.
func (ca *ContentAnalyzer) calculateKeywordDF(noteTexts map[string]string) map[string]int {
	df := make(map[string]int)

	for _, text := range noteTexts {
		words := ca.extractWords(text)
		seen := make(map[string]bool)

		for _, word := range words {
			if !seen[word] && !ca.stopWords[word] && len(word) > 2 {
				df[word]++
				seen[word] = true
			}
		}
	}

	return df
}

// calculatePhraseDF calculates document frequency for phrases (bigrams/trigrams).
func (ca *ContentAnalyzer) calculatePhraseDF(noteTexts map[string]string) map[string]int {
	df := make(map[string]int)

	for _, text := range noteTexts {
		words := ca.extractWords(text)
		seen := make(map[string]bool)

		// Extract bigrams
		for i := 0; i < len(words)-1; i++ {
			if ca.stopWords[words[i]] {
				continue
			}
			phrase := words[i] + " " + words[i+1]
			if !seen[phrase] && len(phrase) > 5 {
				df[phrase]++
				seen[phrase] = true
			}
		}

		// Extract trigrams
		for i := 0; i < len(words)-2; i++ {
			if ca.stopWords[words[i]] {
				continue
			}
			phrase := words[i] + " " + words[i+1] + " " + words[i+2]
			if !seen[phrase] && len(phrase) > 8 {
				df[phrase]++
				seen[phrase] = true
			}
		}
	}

	return df
}

// calculateTFIDF calculates TF-IDF scores for keywords.
func (ca *ContentAnalyzer) calculateTFIDF(noteTexts map[string]string, df map[string]int) map[string]map[string]float64 {
	scores := make(map[string]map[string]float64)
	numDocs := float64(len(noteTexts))

	for uuid, text := range noteTexts {
		words := ca.extractWords(text)
		tf := make(map[string]int)

		// Calculate term frequency
		for _, word := range words {
			if !ca.stopWords[word] && len(word) > 2 {
				tf[word]++
			}
		}

		// Calculate TF-IDF
		scores[uuid] = make(map[string]float64)
		for word, freq := range tf {
			if df[word] < ca.minKeywordFreq {
				continue
			}

			termFreq := float64(freq) / float64(len(words))
			inverseDocFreq := math.Log(numDocs / float64(df[word]))
			scores[uuid][word] = termFreq * inverseDocFreq
		}
	}

	return scores
}

// calculatePhraseTFIDF calculates TF-IDF scores for phrases.
func (ca *ContentAnalyzer) calculatePhraseTFIDF(noteTexts map[string]string, df map[string]int) map[string]map[string]float64 {
	scores := make(map[string]map[string]float64)
	numDocs := float64(len(noteTexts))

	for uuid, text := range noteTexts {
		words := ca.extractWords(text)
		tf := make(map[string]int)

		// Calculate phrase frequency
		for i := 0; i < len(words)-1; i++ {
			if ca.stopWords[words[i]] {
				continue
			}
			phrase := words[i] + " " + words[i+1]
			if len(phrase) > 5 {
				tf[phrase]++
			}
		}

		for i := 0; i < len(words)-2; i++ {
			if ca.stopWords[words[i]] {
				continue
			}
			phrase := words[i] + " " + words[i+1] + " " + words[i+2]
			if len(phrase) > 8 {
				tf[phrase]++
			}
		}

		// Calculate TF-IDF
		scores[uuid] = make(map[string]float64)
		for phrase, freq := range tf {
			if df[phrase] < ca.minPhraseFreq {
				continue
			}

			termFreq := float64(freq) / float64(len(words))
			inverseDocFreq := math.Log(numDocs / float64(df[phrase]))
			scores[uuid][phrase] = termFreq * inverseDocFreq
		}
	}

	return scores
}

// identifyThemes clusters notes by their top keywords and phrases.
func (ca *ContentAnalyzer) identifyThemes(keywordScores, phraseScores map[string]map[string]float64, noteTexts map[string]string) []ContentTheme {
	// Aggregate scores across all notes
	globalKeywords := make(map[string]float64)
	globalPhrases := make(map[string]float64)

	for uuid := range noteTexts {
		for word, score := range keywordScores[uuid] {
			globalKeywords[word] += score
		}
		for phrase, score := range phraseScores[uuid] {
			globalPhrases[phrase] += score
		}
	}

	// Create themes from top terms
	themes := make(map[string]*ContentTheme)

	// Group notes by their top keywords
	for uuid, scores := range keywordScores {
		topTerms := ca.getTopTerms(scores, 5)

		for _, term := range topTerms {
			// Only create themes for globally significant terms
			if _, exists := globalKeywords[term]; !exists || globalKeywords[term] < 1.0 {
				continue
			}

			themeName := toTitleCase(term)
			if theme, exists := themes[themeName]; exists {
				theme.RelatedNotes = append(theme.RelatedNotes, uuid)
				theme.NoteCount++
			} else {
				themes[themeName] = &ContentTheme{
					Name:         themeName,
					Keywords:     []string{term},
					Phrases:      []string{},
					NoteCount:    1,
					Relevance:    globalKeywords[term],
					RelatedNotes: []string{uuid},
				}
			}
		}
	}

	// Add phrases to themes
	for uuid, scores := range phraseScores {
		topPhrases := ca.getTopTerms(scores, 3)

		for _, phrase := range topPhrases {
			if _, exists := globalPhrases[phrase]; !exists || globalPhrases[phrase] < 1.0 {
				continue
			}

			// Find which theme this note belongs to
			for _, theme := range themes {
				for _, noteUUID := range theme.RelatedNotes {
					if noteUUID == uuid {
						if !contains(theme.Phrases, phrase) {
							theme.Phrases = append(theme.Phrases, phrase)
						}
						break
					}
				}
			}
		}
	}

	// Convert to slice and filter by minimum notes
	var result []ContentTheme
	for _, theme := range themes {
		if theme.NoteCount >= ca.minThemeNotes {
			result = append(result, *theme)
		}
	}

	// Sort by relevance
	sort.Slice(result, func(i, j int) bool {
		return result[i].Relevance > result[j].Relevance
	})

	// Limit to top 20 themes
	if len(result) > 20 {
		result = result[:20]
	}

	return result
}

// getTopTerms returns the top N terms by score.
func (ca *ContentAnalyzer) getTopTerms(scores map[string]float64, n int) []string {
	type termScore struct {
		term  string
		score float64
	}

	var sorted []termScore
	for term, score := range scores {
		sorted = append(sorted, termScore{term, score})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score > sorted[j].score
	})

	limit := n
	if len(sorted) < limit {
		limit = len(sorted)
	}

	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		result[i] = sorted[i].term
	}

	return result
}

// extractWords tokenizes text into words.
func (ca *ContentAnalyzer) extractWords(text string) []string {
	// Remove markdown syntax
	text = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile("[ *_~`]").ReplaceAllString(text, "")

	// Split on non-letter characters
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	// Clean and lowercase
	var result []string
	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		if len(word) > 0 {
			result = append(result, word)
		}
	}

	return result
}

// buildStopWords returns common English stop words.
func buildStopWords() map[string]bool {
	words := []string{
		"a", "about", "above", "after", "again", "against", "all", "am", "an", "and", "any", "are", "aren't",
		"as", "at", "be", "because", "been", "before", "being", "below", "between", "both", "but", "by",
		"can", "can't", "cannot", "could", "couldn't", "did", "didn't", "do", "does", "doesn't", "doing",
		"don't", "down", "during", "each", "few", "for", "from", "further", "had", "hadn't", "has", "hasn't",
		"have", "haven't", "having", "he", "he'd", "he'll", "he's", "her", "here", "here's", "hers", "herself",
		"him", "himself", "his", "how", "how's", "i", "i'd", "i'll", "i'm", "i've", "if", "in", "into", "is",
		"isn't", "it", "it's", "its", "itself", "let's", "me", "more", "most", "mustn't", "my", "myself",
		"no", "nor", "not", "of", "off", "on", "once", "only", "or", "other", "ought", "our", "ours",
		"ourselves", "out", "over", "own", "same", "shan't", "she", "she'd", "she'll", "she's", "should",
		"shouldn't", "so", "some", "such", "than", "that", "that's", "the", "their", "theirs", "them",
		"themselves", "then", "there", "there's", "these", "they", "they'd", "they'll", "they're", "they've",
		"this", "those", "through", "to", "too", "under", "until", "up", "very", "was", "wasn't", "we",
		"we'd", "we'll", "we're", "we've", "were", "weren't", "what", "what's", "when", "when's", "where",
		"where's", "which", "while", "who", "who's", "whom", "why", "why's", "with", "won't", "would",
		"wouldn't", "you", "you'd", "you'll", "you're", "you've", "your", "yours", "yourself", "yourselves",
		"also", "just", "like", "etc", "via", "using", "use", "used", "one", "two", "three", "new", "get",
		"make", "made", "way", "see", "know", "think", "take", "need", "want", "back", "go", "going", "come",
	}

	stopWords := make(map[string]bool)
	for _, word := range words {
		stopWords[word] = true
	}

	return stopWords
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
