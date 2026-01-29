package sncli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	"google.golang.org/api/option"
)

// OrganizeConfig holds configuration for the organize operation.
type OrganizeConfig struct {
	Session   *cache.Session
	GeminiKey string
	Since     string   // Date filter (RFC3339 format)
	Until     string   // Date filter (RFC3339 format)
	UUIDs     []string // Specific note UUIDs to process
	Titles    []string // Specific note titles to process
	AutoApply bool     // Skip confirmation prompt
	Debug     bool
}

// OrganizeOutput contains results of the organize operation.
type OrganizeOutput struct {
	Changes      []ProposedChange
	Applied      bool
	NotesUpdated int
	TagsCreated  []string
}

// ProposedChange represents a single note's proposed modifications.
type ProposedChange struct {
	NoteUUID     string
	NoteTitle    string
	NewTitle     string
	TitleChanged bool
	ExistingTags []string
	ProposedTags []string
	TagsChanged  bool
	Reason       string // AI's reasoning for changes
}

// GeminiResponse is the structured response from Gemini API.
type GeminiResponse struct {
	Title     string   `json:"title"`      // Proposed title (empty if no change)
	Tags      []string `json:"tags"`       // Proposed tags
	Reasoning string   `json:"reasoning"`  // Explanation of changes
	CreateNew []string `json:"create_new"` // New tags that need creation
}

// Run executes the organize operation and returns proposed changes.
func (i *OrganizeConfig) Run() (OrganizeOutput, error) {
	output := OrganizeOutput{}

	// 1. Validate Gemini API key
	if i.GeminiKey == "" {
		return output, errors.New("GEMINI_API_KEY not provided")
	}

	// 2. Initialize Gemini client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(i.GeminiKey))
	if err != nil {
		return output, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")

	// 3. Sync and get all existing tags
	syncInput := cache.SyncInput{Session: i.Session}
	so, err := Sync(syncInput, true)
	if err != nil {
		return output, err
	}

	var allPersistedItems cache.Items
	if err = so.DB.All(&allPersistedItems); err != nil {
		return output, err
	}

	gItems, err := allPersistedItems.ToItems(i.Session)
	if err != nil {
		return output, err
	}

	// Extract all existing tags
	var existingTags []items.Tag
	var tagTitles []string
	for _, item := range gItems {
		if !item.IsDeleted() && item.GetContentType() == common.SNItemTypeTag {
			tag := item.(*items.Tag)
			existingTags = append(existingTags, *tag)
			tagTitles = append(tagTitles, tag.Content.Title)
		}
	}

	// 4. Build note filters based on user criteria
	filters := buildNoteFilters(i)

	// 5. Get notes to process
	noteConfig := GetNoteConfig{
		Session: i.Session,
		Filters: filters,
		Debug:   i.Debug,
	}

	noteItems, err := noteConfig.Run()
	if err != nil {
		return output, err
	}

	notes := noteItems.Notes()
	if len(notes) == 0 {
		return output, errors.New("no notes found matching criteria")
	}

	// 6. Process each note with Gemini
	for _, note := range notes {
		// Get existing tags for this note
		noteTags := getTagsForNote(note, existingTags)

		// Call Gemini API
		change, err := analyzeNoteWithGemini(ctx, model, note, noteTags, tagTitles, i.Debug)
		if err != nil {
			if i.Debug {
				fmt.Printf("Error analyzing note %s: %v\n", note.UUID, err)
			}
			continue // Skip this note but continue with others
		}

		if change.TitleChanged || change.TagsChanged {
			output.Changes = append(output.Changes, change)
		}
	}

	return output, nil
}

// buildNoteFilters converts OrganizeConfig criteria to items.ItemFilters.
func buildNoteFilters(config *OrganizeConfig) items.ItemFilters {
	filters := items.ItemFilters{
		MatchAny: false,
		Filters: []items.Filter{
			{Type: common.SNItemTypeNote},
		},
	}

	// Add UUID filters
	if len(config.UUIDs) > 0 {
		for _, uuid := range config.UUIDs {
			filters.Filters = append(filters.Filters, items.Filter{
				Type:       common.SNItemTypeNote,
				Key:        "uuid",
				Comparison: "==",
				Value:      uuid,
			})
		}
		filters.MatchAny = true
	}

	// Add title filters
	if len(config.Titles) > 0 {
		for _, title := range config.Titles {
			filters.Filters = append(filters.Filters, items.Filter{
				Type:       common.SNItemTypeNote,
				Key:        "Title",
				Comparison: "contains",
				Value:      title,
			})
		}
		filters.MatchAny = true
	}

	// Add date filters
	if config.Since != "" {
		filters.Filters = append(filters.Filters, items.Filter{
			Type:       common.SNItemTypeNote,
			Key:        "created_at",
			Comparison: ">=",
			Value:      config.Since,
		})
	}

	if config.Until != "" {
		filters.Filters = append(filters.Filters, items.Filter{
			Type:       common.SNItemTypeNote,
			Key:        "created_at",
			Comparison: "<=",
			Value:      config.Until,
		})
	}

	return filters
}

// getTagsForNote returns the tags assigned to a note.
func getTagsForNote(note items.Note, allTags []items.Tag) []string {
	var tagTitles []string

	for _, tag := range allTags {
		for _, ref := range tag.Content.References() {
			if ref.UUID == note.UUID && ref.ContentType == common.SNItemTypeNote {
				tagTitles = append(tagTitles, tag.Content.Title)
				break
			}
		}
	}

	return tagTitles
}

// analyzeNoteWithGemini calls Gemini API to analyze a note and suggest improvements.
func analyzeNoteWithGemini(
	ctx context.Context,
	model *genai.GenerativeModel,
	note items.Note,
	currentTags []string,
	allExistingTags []string,
	debug bool,
) (ProposedChange, error) {
	change := ProposedChange{
		NoteUUID:     note.UUID,
		NoteTitle:    note.Content.Title,
		ExistingTags: currentTags,
	}

	// Build prompt
	prompt := buildGeminiPrompt(note, currentTags, allExistingTags)

	if debug {
		fmt.Printf("\n--- Analyzing note: %s ---\n", note.Content.Title)
		fmt.Printf("Prompt: %s\n", prompt)
	}

	// Configure JSON response
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"title": {
				Type:        genai.TypeString,
				Description: "Improved title (empty string if current title is good)",
			},
			"tags": {
				Type:        genai.TypeArray,
				Items:       &genai.Schema{Type: genai.TypeString},
				Description: "Array of tag titles to apply",
			},
			"create_new": {
				Type:        genai.TypeArray,
				Items:       &genai.Schema{Type: genai.TypeString},
				Description: "Array of new tag titles that don't exist yet",
			},
			"reasoning": {
				Type:        genai.TypeString,
				Description: "Brief explanation of changes",
			},
		},
		Required: []string{"tags", "create_new", "reasoning"},
	}

	// Call Gemini API
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return change, fmt.Errorf("Gemini API error: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return change, errors.New("empty response from Gemini")
	}

	// Parse JSON response
	var geminiResp GeminiResponse
	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	if debug {
		fmt.Printf("Gemini response: %s\n", responseText)
	}

	if err := json.Unmarshal([]byte(responseText), &geminiResp); err != nil {
		return change, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	// Process response
	if geminiResp.Title != "" && geminiResp.Title != note.Content.Title {
		change.NewTitle = geminiResp.Title
		change.TitleChanged = true
	}

	if !stringSlicesEqual(currentTags, geminiResp.Tags) {
		change.ProposedTags = geminiResp.Tags
		change.TagsChanged = true
	} else {
		change.ProposedTags = currentTags
	}

	change.Reason = geminiResp.Reasoning

	// Add small delay to respect rate limits
	const rateLimitDelay = 500
	time.Sleep(rateLimitDelay * time.Millisecond)

	return change, nil
}

// buildGeminiPrompt creates the prompt for Gemini API.
func buildGeminiPrompt(note items.Note, currentTags, allTags []string) string {
	prompt := `You are an intelligent note organization assistant. Analyze this note and suggest improvements.

EXISTING TAGS IN SYSTEM:
%s

CURRENT NOTE INFORMATION:
Title: %s
Content:
%s

Current Tags: %s

INSTRUCTIONS:
1. Review the note content carefully
2. Suggest relevant tags from existing tags when possible
3. Only create NEW tags if existing ones don't capture important topics
4. Improve the title ONLY if it's vague, generic, or unclear
5. Keep titles concise (under 60 characters)
6. Prefer existing tags to maintain organization consistency

Return your analysis as JSON with this structure:
{
  "title": "Improved title (empty string if current is good)",
  "tags": ["tag1", "tag2", "tag3"],
  "create_new": ["new_tag1", "new_tag2"],
  "reasoning": "Brief explanation of your changes"
}`

	existingTagsList := strings.Join(allTags, ", ")
	if existingTagsList == "" {
		existingTagsList = "None"
	}

	currentTagsList := strings.Join(currentTags, ", ")
	if currentTagsList == "" {
		currentTagsList = "None"
	}

	// Truncate note content if too long (Gemini has token limits)
	const maxContentLength = 4000
	noteContent := note.Content.Text
	if len(noteContent) > maxContentLength {
		noteContent = noteContent[:maxContentLength] + "\n\n[Content truncated...]"
	}

	return fmt.Sprintf(prompt, existingTagsList, note.Content.Title, noteContent, currentTagsList)
}

// ApplyOrganizeChanges applies the confirmed changes to notes and tags.
func ApplyOrganizeChanges(session *cache.Session, changes []ProposedChange, debug bool) error {
	// 1. Sync to get latest data
	syncInput := cache.SyncInput{Session: session}
	so, err := Sync(syncInput, true)
	if err != nil {
		return err
	}

	// 2. Get all items
	var allPersistedItems cache.Items
	if err = so.DB.All(&allPersistedItems); err != nil {
		return err
	}

	gItems, err := allPersistedItems.ToItems(session)
	if err != nil {
		return err
	}

	// 3. Collect all new tags that need creation
	var newTagsNeeded []string
	newTagsMap := make(map[string]bool)

	for _, change := range changes {
		if change.TagsChanged {
			for _, tag := range change.ProposedTags {
				if !tagExistsInItems(tag, gItems) && !newTagsMap[tag] {
					newTagsNeeded = append(newTagsNeeded, tag)
					newTagsMap[tag] = true
				}
			}
		}
	}

	// 4. Create new tags if needed
	if len(newTagsNeeded) > 0 {
		ati := addTagsInput{
			session:   session,
			tagTitles: newTagsNeeded,
		}
		if _, err := addTags(ati); err != nil {
			return fmt.Errorf("failed to create new tags: %w", err)
		}

		// Re-sync to get newly created tags
		so, err = Sync(syncInput, true)
		if err != nil {
			return err
		}

		allPersistedItems = cache.Items{}
		if err = so.DB.All(&allPersistedItems); err != nil {
			return err
		}

		gItems, err = allPersistedItems.ToItems(session)
		if err != nil {
			return err
		}
	}

	// 5. Update notes (titles)
	var notesToUpdate items.Notes
	for _, change := range changes {
		if change.TitleChanged {
			// Find the note in gItems
			for _, item := range gItems {
				if item.GetContentType() == common.SNItemTypeNote {
					note := item.(*items.Note)
					if note.UUID == change.NoteUUID {
						note.Content.Title = change.NewTitle
						note.Content.SetUpdateTime(time.Now().UTC())
						notesToUpdate = append(notesToUpdate, *note)
						break
					}
				}
			}
		}
	}

	if len(notesToUpdate) > 0 {
		if err = cache.SaveNotes(session, so.DB, notesToUpdate, false); err != nil {
			return fmt.Errorf("failed to save updated notes: %w", err)
		}
	}

	// 6. Update tag references
	var tagsToUpdate items.Tags
	allTags := extractTags(gItems)

	for _, change := range changes {
		if change.TagsChanged {
			// Update tag references for this note
			typeUUIDs := map[string][]string{
				common.SNItemTypeNote: {change.NoteUUID},
			}

			for _, tagTitle := range change.ProposedTags {
				for i, tag := range allTags {
					if tag.Content.Title == tagTitle {
						updatedTag, changed := upsertTagReferences(allTags[i], typeUUIDs)
						if changed {
							tagsToUpdate = append(tagsToUpdate, updatedTag)
						}
						break
					}
				}
			}
		}
	}

	if len(tagsToUpdate) > 0 {
		if err = cache.SaveTags(so.DB, session, tagsToUpdate, true); err != nil {
			return fmt.Errorf("failed to save updated tags: %w", err)
		}
	}

	// 7. Final sync
	if _, err = Sync(syncInput, true); err != nil {
		return err
	}

	return nil
}

// Helper functions

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aMap := make(map[string]bool)
	for _, s := range a {
		aMap[s] = true
	}
	for _, s := range b {
		if !aMap[s] {
			return false
		}
	}
	return true
}

func tagExistsInItems(title string, allItems items.Items) bool {
	for _, item := range allItems {
		if item.GetContentType() == common.SNItemTypeTag {
			tag := item.(*items.Tag)
			if tag.Content.Title == title {
				return true
			}
		}
	}
	return false
}

func extractTags(gItems items.Items) items.Tags {
	var tags items.Tags
	for _, item := range gItems {
		if !item.IsDeleted() && item.GetContentType() == common.SNItemTypeTag {
			tag := item.(*items.Tag)
			tags = append(tags, *tag)
		}
	}
	return tags
}
