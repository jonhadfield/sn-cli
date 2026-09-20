package sncli

import (
	"testing"

	"github.com/jonhadfield/gosn-v2/cache"
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
				Session:   &cache.Session{},
				OutputDir: "/tmp/test",
			},
			wantErr: true,
		},
		{
			name: "missing output dir",
			config: MigrateConfig{
				Session:  &cache.Session{},
				Provider: "obsidian",
			},
			wantErr: true,
		},
		{
			name: "invalid MOC style",
			config: MigrateConfig{
				Session:      &cache.Session{},
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
				Session:      &cache.Session{},
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
	tag1, _ := items.NewTag("work", nil)
	tag2, _ := items.NewTag("personal", nil)

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
	tag1, _ := items.NewTag("work", nil)
	tag2, _ := items.NewTag("personal", nil)
	tag3, _ := items.NewTag("rarely-used", nil)

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
	tag1, _ := items.NewTag("work", nil)

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

// nestedTagFixture builds work > work-clients > work-clients-acme, each with
// one note, plus an untagged-parent tag "reading" with one note.
func nestedTagFixture(t *testing.T) items.Items {
	t.Helper()

	work, err := items.NewTag("work", nil)
	require.NoError(t, err)

	clients, err := items.NewTag("clients", nil)
	require.NoError(t, err)
	clients.Content.UpsertReferences(items.ItemReferences{
		{UUID: work.UUID, ContentType: "Tag", ReferenceType: "TagToParentTag"},
	})

	acme, err := items.NewTag("acme", nil)
	require.NoError(t, err)
	acme.Content.UpsertReferences(items.ItemReferences{
		{UUID: clients.UUID, ContentType: "Tag", ReferenceType: "TagToParentTag"},
	})

	reading, err := items.NewTag("reading", nil)
	require.NoError(t, err)

	allItems := items.Items{&work, &clients, &acme, &reading}

	for tagUUID, title := range map[string]string{
		work.UUID:    "Work Note",
		clients.UUID: "Clients Note",
		acme.UUID:    "Acme Note",
		reading.UUID: "Reading Note",
	} {
		note, nErr := items.NewNote(title, "Content", nil)
		require.NoError(t, nErr)
		note.Content.UpsertReferences(items.ItemReferences{
			{UUID: tagUUID, ContentType: "Tag"},
		})
		allItems = append(allItems, &note)
	}

	return allItems
}

func mocByTitle(mocs []MOCFile, title string) (MOCFile, bool) {
	for _, moc := range mocs {
		if moc.Title == title {
			return moc, true
		}
	}

	return MOCFile{}, false
}

func TestMOCBuilder_HierarchicalFollowsTagTree(t *testing.T) {
	builder := NewMOCBuilder(nestedTagFixture(t), MOCConfig{
		Style:    MOCStyleHierarchical,
		MaxDepth: 3,
	})

	mocs, err := builder.Generate()
	require.NoError(t, err)

	// Home lists only the roots, not the nested tags
	home, ok := mocByTitle(mocs, "Home")
	require.True(t, ok)
	assert.Contains(t, home.Content, "[[Work MOC]]")
	assert.Contains(t, home.Content, "[[Reading MOC]]")
	assert.NotContains(t, home.Content, "[[Clients MOC]]")

	// a parent links to its child's MOC, and counts the whole subtree
	work, ok := mocByTitle(mocs, "Work MOC")
	require.True(t, ok)
	assert.Contains(t, work.Content, "Sub-categories")
	assert.Contains(t, work.Content, "[[Clients MOC]]")
	assert.Contains(t, work.Content, "[[Work Note]]")

	// every level of the tree got its own MOC
	for _, title := range []string{"Clients MOC", "Acme MOC"} {
		_, found := mocByTitle(mocs, title)
		assert.True(t, found, "expected a MOC for %s", title)
	}
}

func TestMOCBuilder_HierarchicalRespectsDepth(t *testing.T) {
	builder := NewMOCBuilder(nestedTagFixture(t), MOCConfig{
		Style:    MOCStyleHierarchical,
		MaxDepth: 2,
	})

	mocs, err := builder.Generate()
	require.NoError(t, err)

	// depth 2 stops at clients, so acme gets no MOC of its own
	_, found := mocByTitle(mocs, "Acme MOC")
	assert.False(t, found, "acme is below the depth limit and should not have a MOC")

	// but its note is still reachable, listed on the deepest MOC
	clients, ok := mocByTitle(mocs, "Clients MOC")
	require.True(t, ok)
	assert.Contains(t, clients.Content, "[[Acme Note]]")
	assert.NotContains(t, clients.Content, "[[Acme MOC]]")
}

func TestMOCBuilder_HierarchicalFallsBackWithoutTree(t *testing.T) {
	tag, err := items.NewTag("work", nil)
	require.NoError(t, err)

	allItems := items.Items{&tag}

	for i := 0; i < 3; i++ {
		note, nErr := items.NewNote("Work Note", "Content", nil)
		require.NoError(t, nErr)
		note.Content.UpsertReferences(items.ItemReferences{
			{UUID: tag.UUID, ContentType: "Tag"},
		})
		allItems = append(allItems, &note)
	}

	builder := NewMOCBuilder(allItems, MOCConfig{
		Style:          MOCStyleHierarchical,
		MaxDepth:       3,
		MinNotesPerMOC: 1,
	})

	mocs, err := builder.Generate()
	require.NoError(t, err)

	home, ok := mocByTitle(mocs, "Home")
	require.True(t, ok)
	assert.Contains(t, home.Content, "[[work MOC]]")
}

func TestClassifyPARATag(t *testing.T) {
	cases := map[string]string{
		"project-apollo": "Projects",
		"sprint-14":      "Projects",
		"health":         "Areas",
		"finance":        "Areas",
		"archived-2024":  "Archive",
		"done":           "Archive",
		"reading":        "Resources",
		"kubernetes":     "Resources", // no keyword match falls back to Resources
	}

	for tag, want := range cases {
		t.Run(tag, func(t *testing.T) {
			assert.Equal(t, want, classifyPARATag(tag))
		})
	}
}

func TestMOCBuilder_PARASortsTagsIntoBuckets(t *testing.T) {
	var allItems items.Items

	for _, title := range []string{"project-apollo", "health", "archived-2024", "kubernetes"} {
		tag, err := items.NewTag(title, nil)
		require.NoError(t, err)

		allItems = append(allItems, &tag)

		note, nErr := items.NewNote(title+" note", "Content", nil)
		require.NoError(t, nErr)
		note.Content.UpsertReferences(items.ItemReferences{
			{UUID: tag.UUID, ContentType: "Tag"},
		})
		allItems = append(allItems, &note)
	}

	builder := NewMOCBuilder(allItems, MOCConfig{Style: MOCStylePARA})

	mocs, err := builder.Generate()
	require.NoError(t, err)

	for _, title := range []string{"Projects MOC", "Areas MOC", "Archive MOC", "Resources MOC"} {
		_, found := mocByTitle(mocs, title)
		assert.True(t, found, "expected %s", title)
	}

	projects, ok := mocByTitle(mocs, "Projects MOC")
	require.True(t, ok)
	assert.Contains(t, projects.Content, "[[project-apollo note]]")
	assert.NotContains(t, projects.Content, "[[health note]]")

	resources, ok := mocByTitle(mocs, "Resources MOC")
	require.True(t, ok)
	assert.Contains(t, resources.Content, "[[kubernetes note]]")
}

func TestMOCBuilder_AutoPicksHierarchicalWhenTreeExists(t *testing.T) {
	builder := NewMOCBuilder(nestedTagFixture(t), MOCConfig{
		Style:    MOCStyleAuto,
		MaxDepth: 3,
	})

	mocs, err := builder.Generate()
	require.NoError(t, err)

	work, ok := mocByTitle(mocs, "Work MOC")
	require.True(t, ok)
	assert.Contains(t, work.Content, "Sub-categories")
}
