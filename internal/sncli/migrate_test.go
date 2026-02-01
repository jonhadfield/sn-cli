package sncli

import (
	"testing"

	"github.com/jonhadfield/gosn-v2/items"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  MigrateConfig
		wantErr bool
	}{
		{
			name: "missing session",
			config: MigrateConfig{
				Provider:  "obsidian",
				OutputDir: "/tmp/test",
			},
			wantErr: true,
		},
		{
			name: "missing provider",
			config: MigrateConfig{
				Session:   &struct{}{}, // dummy session
				OutputDir: "/tmp/test",
			},
			wantErr: true,
		},
		{
			name: "missing output dir",
			config: MigrateConfig{
				Session:  &struct{}{},
				Provider: "obsidian",
			},
			wantErr: true,
		},
		{
			name: "invalid MOC style",
			config: MigrateConfig{
				Session:      &struct{}{},
				Provider:     "obsidian",
				OutputDir:    "/tmp/test",
				GenerateMOCs: true,
				MOCStyle:     "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid MOC depth",
			config: MigrateConfig{
				Session:      &struct{}{},
				Provider:     "obsidian",
				OutputDir:    "/tmp/test",
				GenerateMOCs: true,
				MOCStyle:     MOCStyleFlat,
				MOCDepth:     0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestObsidianExporter_SanitizeFilename(t *testing.T) {
	exporter := NewObsidianExporter("/tmp/test")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid filename",
			input:    "My Note Title",
			expected: "My Note Title",
		},
		{
			name:     "remove invalid chars",
			input:    "My:Note/Title",
			expected: "My-Note-Title",
		},
		{
			name:     "multiple spaces",
			input:    "My   Note   Title",
			expected: "My Note Title",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "untitled",
		},
		{
			name:     "long filename",
			input:    string(make([]byte, 250)),
			expected: string(make([]byte, 200)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.sanitizeFilename(tt.input)
			if tt.name == "long filename" {
				assert.LessOrEqual(t, len(result), 200)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestExtractNoteTags(t *testing.T) {
	// Create test notes and tags
	tag1, _ := items.NewTag("work")
	tag2, _ := items.NewTag("personal")

	note, _ := items.NewNote("Test Note", "Test content", nil)
	note.Content.UpsertReferences(items.ItemReferences{
		{UUID: tag1.UUID, ContentType: "Tag"},
		{UUID: tag2.UUID, ContentType: "Tag"},
	})

	allItems := items.Items{&tag1, &tag2, &note}

	tags := extractNoteTags(&note, allItems)

	assert.Len(t, tags, 2)
	assert.Contains(t, tags, "work")
	assert.Contains(t, tags, "personal")
}

func TestMOCBuilder_IdentifyTopLevelTags(t *testing.T) {
	// Create test data
	tag1, _ := items.NewTag("work")
	tag2, _ := items.NewTag("personal")
	tag3, _ := items.NewTag("rarely-used")

	// Create 10 notes with work tag, 5 with personal, 1 with rarely-used
	var allItems items.Items
	allItems = append(allItems, &tag1, &tag2, &tag3)

	for i := 0; i < 10; i++ {
		note, _ := items.NewNote("Work Note", "Content", nil)
		note.Content.UpsertReferences(items.ItemReferences{
			{UUID: tag1.UUID, ContentType: "Tag"},
		})
		allItems = append(allItems, &note)
	}

	for i := 0; i < 5; i++ {
		note, _ := items.NewNote("Personal Note", "Content", nil)
		note.Content.UpsertReferences(items.ItemReferences{
			{UUID: tag2.UUID, ContentType: "Tag"},
		})
		allItems = append(allItems, &note)
	}

	note, _ := items.NewNote("Rare Note", "Content", nil)
	note.Content.UpsertReferences(items.ItemReferences{
		{UUID: tag3.UUID, ContentType: "Tag"},
	})
	allItems = append(allItems, &note)

	// Build MOC
	config := MOCConfig{
		MinNotesPerMOC: 3,
	}
	builder := NewMOCBuilder(allItems, config)

	topTags := builder.identifyTopLevelTags()

	// Should include work and personal, but not rarely-used
	assert.Contains(t, topTags, "work")
	assert.Contains(t, topTags, "personal")
	assert.NotContains(t, topTags, "rarely-used")
}

func TestEscapeYAMLString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no special chars",
			input:    "Simple Title",
			expected: "Simple Title",
		},
		{
			name:     "with quotes",
			input:    `Title with "quotes"`,
			expected: `Title with \"quotes\"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeYAMLString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMOCBuilder_Generate(t *testing.T) {
	// Create minimal test data
	tag1, _ := items.NewTag("work")

	var allItems items.Items
	allItems = append(allItems, &tag1)

	for i := 0; i < 5; i++ {
		note, _ := items.NewNote("Work Note", "Content", nil)
		note.Content.UpsertReferences(items.ItemReferences{
			{UUID: tag1.UUID, ContentType: "Tag"},
		})
		allItems = append(allItems, &note)
	}

	config := MOCConfig{
		Style:          MOCStyleFlat,
		MinNotesPerMOC: 3,
		IncludeStats:   true,
		IncludeRecent:  true,
		RecentCount:    5,
	}

	builder := NewMOCBuilder(allItems, config)
	mocs, err := builder.Generate()

	require.NoError(t, err)
	require.NotEmpty(t, mocs)

	// Should have at least Home MOC
	assert.GreaterOrEqual(t, len(mocs), 1)

	// First MOC should be Home
	assert.Equal(t, "Home.md", mocs[0].Filename)
	assert.Equal(t, "Home", mocs[0].Title)
	assert.Contains(t, mocs[0].Content, "# 🏠 Home")
}
