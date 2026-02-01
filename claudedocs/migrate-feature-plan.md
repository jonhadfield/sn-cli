# Migration Feature Implementation Plan

## Overview
Add a `migrate` command to sn-cli that exports Standard Notes content to other note-taking applications with intelligent organization and MOC (Maps of Content) generation.

## Target Applications (Phased)

### Phase 1: Obsidian
Primary target due to:
- Pure markdown format
- Excellent wikilink support
- Strong community
- Local-first architecture
- Advanced linking/organization features

### Phase 2: Other Providers
- Logseq (outliner-based, markdown)
- Notion (API-based export)
- Joplin (markdown with resources)
- Bear (markdown with tags)

## Architecture

### Command Structure
```bash
sn migrate <provider> [options]

# Examples
sn migrate obsidian --output ./my-vault
sn migrate obsidian --output ./vault --moc --moc-style hierarchical
sn migrate obsidian --output ./vault --include-tags work,personal
sn migrate obsidian --output ./vault --dry-run
```

### File Structure
```
internal/sncli/
  ├── migrate.go              # Core migration logic
  ├── migrate_obsidian.go     # Obsidian-specific implementation
  ├── migrate_moc.go          # MOC generation logic
  ├── migrate_analyzer.go     # Content analysis for smart MOCs
  └── migrate_test.go         # Tests

cmd/sncli/
  └── migrate.go              # CLI command handler
```

## Obsidian Export Format

### Note Format
```markdown
---
title: "Meeting Notes - Q1 Planning"
tags: [work, meetings, planning, q1-2024]
created: 2024-01-15T10:30:00Z
updated: 2024-01-15T14:20:00Z
uuid: abc123-def456-ghi789
source: standard-notes
---

# Meeting Notes - Q1 Planning

## Attendees
- Alice
- Bob

## Action Items
- [[Project Alpha]] - Review requirements
- [[Budget Planning]] - Finalize Q1 budget
- Follow up with [[Team Lead]]

## Notes
Discussion about quarterly objectives...

## Related Notes
- [[Previous Meeting - Q4 Retrospective]]
- [[OKR Planning]]

#work #meetings #planning
```

### Folder Structure Options

**Option 1: Flat with MOCs (Recommended)**
```
obsidian-vault/
├── Home.md                    # Main MOC
├── Projects.md                # Projects MOC
├── Work.md                    # Work MOC
├── Learning.md                # Learning MOC
├── Meeting Notes - Q1.md
├── Project Alpha Requirements.md
├── Budget Planning 2024.md
└── ... (all other notes)
```

**Option 2: Hierarchical Folders**
```
obsidian-vault/
├── 00-Home/
│   └── Home.md
├── 01-Work/
│   ├── Work.md (MOC)
│   ├── Meetings/
│   ├── Projects/
│   └── Planning/
├── 02-Personal/
│   ├── Personal.md (MOC)
│   ├── Journal/
│   └── Ideas/
└── 03-Learning/
    └── Learning.md (MOC)
```

**Option 3: PARA Method**
```
obsidian-vault/
├── 1-Projects/
├── 2-Areas/
├── 3-Resources/
├── 4-Archives/
└── 0-MOCs/
    ├── Home.md
    ├── Projects.md
    └── Resources.md
```

## MOC Generation Strategy

### 1. Tag Analysis
```go
type TagAnalysis struct {
    Tag           string
    NoteCount     int
    Frequency     float64
    RelatedTags   []string
    IsTopLevel    bool  // Determined by usage patterns
    Children      []*TagAnalysis
}
```

### 2. MOC Templates

**Home MOC (Entry Point)**
```markdown
---
title: Home
tags: [moc, index]
---

# 🏠 Home

Welcome to your knowledge base!

## 📂 Main Areas

### Work & Projects
- [[Work MOC]] - Professional activities and projects
- [[Projects MOC]] - Active and archived projects
- [[Meetings MOC]] - Meeting notes and agendas

### Learning & Development
- [[Learning MOC]] - Study notes and courses
- [[Books MOC]] - Reading notes and summaries
- [[Skills MOC]] - Skill development tracking

### Personal
- [[Personal MOC]] - Personal notes and thoughts
- [[Ideas MOC]] - Ideas and inspiration
- [[Journal MOC]] - Daily journal entries

## 📊 Quick Stats
- Total Notes: 342
- Total Tags: 45
- Last Updated: 2024-01-15

## 🔍 Recently Updated
- [[Meeting Notes - Q1 Planning]] (2 hours ago)
- [[Project Alpha Requirements]] (1 day ago)
- [[Learning React Hooks]] (3 days ago)
```

**Category MOC (e.g., Work)**
```markdown
---
title: Work
tags: [moc, work]
---

# 💼 Work

## Projects
### Active
- [[Project Alpha]] - Web application redesign
- [[Project Beta]] - API integration
- [[Infrastructure Upgrade]] - Cloud migration

### On Hold
- [[Project Gamma]] - Mobile app development

## Meetings
- [[Weekly Standup Notes]]
- [[Q1 Planning Meeting]]
- [[Architecture Review Sessions]]

## Planning & Strategy
- [[OKR 2024 Q1]]
- [[Team Roadmap]]
- [[Budget Planning]]

## Resources
- [[Company Wiki Links]]
- [[Technical Documentation]]
- [[Team Contacts]]

## Related MOCs
- [[Projects MOC]]
- [[Learning MOC]] (for work-related learning)

---
**Tagged Notes**: #work (124 notes)
**Last Updated**: 2024-01-15
```

**Topic-Based MOC (e.g., Security)**
```markdown
---
title: Security Knowledge Base
tags: [moc, security, learning]
---

# 🔐 Security

Comprehensive security knowledge and best practices.

## Authentication & Authorization
- [[OAuth 2.0]] - OAuth 2.0 protocol overview
- [[PKCE]] - Proof Key for Code Exchange
- [[JWT]] - JSON Web Tokens
- [[mTLS]] - Mutual TLS authentication
- [[RBAC]] - Role-Based Access Control

## Web Security
- [[XSS]] - Cross-Site Scripting attacks and prevention
- [[CSRF]] - Cross-Site Request Forgery
- [[CSP]] - Content Security Policy
- [[CORS]] - Cross-Origin Resource Sharing
- [[SQL Injection]] - Prevention and detection

## Network Security
- [[IPS]] - Intrusion Prevention Systems
- [[IDS]] - Intrusion Detection Systems
- [[WAF]] - Web Application Firewall
- [[DDoS Protection]] - Distributed Denial of Service
- [[VPN]] - Virtual Private Networks

## Cryptography
- [[Encryption Algorithms]]
- [[Hashing Functions]]
- [[Digital Signatures]]
- [[Key Management]]

## Security Standards
- [[OWASP Top 10]]
- [[CIS Benchmarks]]
- [[ISO 27001]]
- [[NIST Framework]]

## Incident Response
- [[Incident Response Plan]]
- [[Security Monitoring]]
- [[Threat Intelligence]]

---
**Related MOCs**: [[Learning MOC]] | [[Work MOC]]
**Total Notes**: 45
**Last Updated**: 2024-01-15
```

### 3. MOC Generation Algorithm

```go
type MOCGenerator struct {
    notes         []Note
    tags          map[string]*TagAnalysis
    relationships map[string][]string  // note links
    style         MOCStyle
}

type MOCStyle string

const (
    MOCStyleFlat         MOCStyle = "flat"         // Single-level MOCs
    MOCStyleHierarchical MOCStyle = "hierarchical" // Nested MOCs
    MOCStylePARA         MOCStyle = "para"         // PARA method
    MOCStyleTopicBased   MOCStyle = "topic"        // Topic/domain-based
    MOCStyleAuto         MOCStyle = "auto"         // AI-powered analysis
)

func (m *MOCGenerator) Generate() ([]MOCFile, error) {
    // 1. Analyze tags and relationships
    m.analyzeTags()
    m.analyzeRelationships()

    // 2. Identify top-level categories
    topLevelTags := m.identifyTopLevelTags()

    // 3. Generate MOC hierarchy
    switch m.style {
    case MOCStyleFlat:
        return m.generateFlatMOCs(topLevelTags)
    case MOCStyleHierarchical:
        return m.generateHierarchicalMOCs(topLevelTags)
    case MOCStylePARA:
        return m.generatePARAMOCs()
    case MOCStyleTopicBased:
        return m.generateTopicMOCs()
    case MOCStyleAuto:
        return m.generateSmartMOCs()
    }
}

func (m *MOCGenerator) identifyTopLevelTags() []string {
    // Algorithm:
    // 1. Tags with >20% of notes are top-level
    // 2. Tags that don't co-occur frequently with other tags
    // 3. Tags that represent broad categories (work, personal, learning, etc.)
}

func (m *MOCGenerator) generateSmartMOCs() ([]MOCFile, error) {
    // Use clustering algorithm to group related notes
    // 1. TF-IDF analysis of note content
    // 2. Cosine similarity for relatedness
    // 3. K-means clustering to identify groups
    // 4. Generate MOCs for each cluster
}
```

## Implementation Details

### 1. Core Migration Logic (`internal/sncli/migrate.go`)

```go
package sncli

import (
    "fmt"
    "path/filepath"
    "time"
)

type MigrateConfig struct {
    Session      *cache.Session
    Provider     string
    OutputDir    string
    GenerateMOCs bool
    MOCStyle     MOCStyle
    MOCDepth     int
    TagFilter    []string
    DryRun       bool
    Debug        bool
}

type MigrationResult struct {
    NotesExported int
    MOCsCreated   int
    TagsProcessed int
    Duration      time.Duration
    OutputPath    string
    Warnings      []string
    Errors        []string
}

func (m *MigrateConfig) Run() (*MigrationResult, error) {
    // 1. Validate configuration
    // 2. Sync to get latest notes
    // 3. Get provider-specific exporter
    // 4. Export notes
    // 5. Generate MOCs if requested
    // 6. Return results
}

type Provider interface {
    Name() string
    Export(notes []Note, config ExportConfig) error
    GenerateMOCs(notes []Note, mocConfig MOCConfig) ([]MOCFile, error)
}
```

### 2. Obsidian Implementation (`internal/sncli/migrate_obsidian.go`)

```go
package sncli

type ObsidianExporter struct {
    outputDir    string
    noteFormat   NoteFormat
    linkStyle    LinkStyle
    tagStyle     TagStyle
    preserveUUID bool
}

type NoteFormat struct {
    UseFrontmatter bool
    IncludeSource  bool
    DateFormat     string
}

type LinkStyle string

const (
    LinkStyleWikilink   LinkStyle = "wikilink"   // [[Note Title]]
    LinkStyleMarkdown   LinkStyle = "markdown"   // [Note Title](note-title.md)
    LinkStyleRelative   LinkStyle = "relative"   // [Note Title](../path/to/note.md)
)

type TagStyle string

const (
    TagStyleFrontmatter TagStyle = "frontmatter" // YAML frontmatter
    TagStyleInline      TagStyle = "inline"      // #tag in text
    TagStyleBoth        TagStyle = "both"        // Both frontmatter and inline
)

func (o *ObsidianExporter) Export(notes []Note, config ExportConfig) error {
    for _, note := range notes {
        // 1. Convert note to Obsidian markdown format
        markdown := o.convertToMarkdown(note)

        // 2. Sanitize filename
        filename := o.sanitizeFilename(note.Title)

        // 3. Write to file
        filepath := filepath.Join(o.outputDir, filename+".md")
        err := os.WriteFile(filepath, []byte(markdown), 0644)
        if err != nil {
            return err
        }
    }

    return nil
}

func (o *ObsidianExporter) convertToMarkdown(note Note) string {
    var sb strings.Builder

    // Add frontmatter
    if o.noteFormat.UseFrontmatter {
        sb.WriteString("---\n")
        sb.WriteString(fmt.Sprintf("title: \"%s\"\n", note.Title))
        sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(note.Tags, ", ")))
        sb.WriteString(fmt.Sprintf("created: %s\n", note.CreatedAt))
        sb.WriteString(fmt.Sprintf("updated: %s\n", note.UpdatedAt))
        if o.preserveUUID {
            sb.WriteString(fmt.Sprintf("uuid: %s\n", note.UUID))
        }
        sb.WriteString("source: standard-notes\n")
        sb.WriteString("---\n\n")
    }

    // Add title as heading
    sb.WriteString(fmt.Sprintf("# %s\n\n", note.Title))

    // Add content
    content := o.convertLinks(note.Content)
    sb.WriteString(content)

    // Add inline tags if configured
    if o.tagStyle == TagStyleInline || o.tagStyle == TagStyleBoth {
        sb.WriteString("\n\n")
        for _, tag := range note.Tags {
            sb.WriteString(fmt.Sprintf("#%s ", tag))
        }
    }

    return sb.String()
}

func (o *ObsidianExporter) convertLinks(content string) string {
    // Convert Standard Notes links/references to Obsidian wikilinks
    // This requires analyzing note references and creating [[links]]
}

func (o *ObsidianExporter) sanitizeFilename(title string) string {
    // Remove invalid filename characters
    // Handle duplicates
    // Ensure cross-platform compatibility
}
```

### 3. MOC Generation (`internal/sncli/migrate_moc.go`)

```go
package sncli

type MOCFile struct {
    Filename string
    Title    string
    Content  string
    Tags     []string
    Order    int  // For sorting MOCs
}

type MOCConfig struct {
    Style          MOCStyle
    MaxDepth       int
    MinNotesPerMOC int
    IncludeStats   bool
    IncludeRecent  bool
    RecentCount    int
}

type MOCBuilder struct {
    notes       []Note
    tags        map[string][]Note
    hierarchy   map[string][]string
    style       MOCStyle
    config      MOCConfig
}

func NewMOCBuilder(notes []Note, config MOCConfig) *MOCBuilder {
    mb := &MOCBuilder{
        notes:     notes,
        tags:      make(map[string][]Note),
        hierarchy: make(map[string][]string),
        style:     config.Style,
        config:    config,
    }

    mb.buildTagIndex()
    mb.analyzeHierarchy()

    return mb
}

func (mb *MOCBuilder) buildTagIndex() {
    // Group notes by tag
    for _, note := range mb.notes {
        for _, tag := range note.Tags {
            mb.tags[tag] = append(mb.tags[tag], note)
        }
    }
}

func (mb *MOCBuilder) analyzeHierarchy() {
    // Detect hierarchical relationships between tags
    // e.g., "work/projects/alpha" creates hierarchy
    for tag := range mb.tags {
        if strings.Contains(tag, "/") || strings.Contains(tag, ".") {
            parts := strings.Split(tag, "/")
            parent := ""
            for i, part := range parts {
                current := strings.Join(parts[:i+1], "/")
                if parent != "" {
                    mb.hierarchy[parent] = append(mb.hierarchy[parent], current)
                }
                parent = current
            }
        }
    }
}

func (mb *MOCBuilder) Generate() ([]MOCFile, error) {
    switch mb.style {
    case MOCStyleFlat:
        return mb.generateFlatMOCs()
    case MOCStyleHierarchical:
        return mb.generateHierarchicalMOCs()
    case MOCStylePARA:
        return mb.generatePARAMOCs()
    case MOCStyleTopicBased:
        return mb.generateTopicMOCs()
    default:
        return mb.generateFlatMOCs()
    }
}

func (mb *MOCBuilder) generateFlatMOCs() ([]MOCFile, error) {
    mocs := []MOCFile{}

    // 1. Generate Home MOC
    homeMOC := mb.createHomeMOC()
    mocs = append(mocs, homeMOC)

    // 2. Generate MOC for each top-level tag
    topLevelTags := mb.identifyTopLevelTags()
    for _, tag := range topLevelTags {
        moc := mb.createTagMOC(tag)
        mocs = append(mocs, moc)
    }

    return mocs, nil
}

func (mb *MOCBuilder) createHomeMOC() MOCFile {
    var sb strings.Builder

    sb.WriteString("---\n")
    sb.WriteString("title: Home\n")
    sb.WriteString("tags: [moc, index]\n")
    sb.WriteString("---\n\n")
    sb.WriteString("# 🏠 Home\n\n")
    sb.WriteString("Welcome to your knowledge base!\n\n")
    sb.WriteString("## 📂 Main Areas\n\n")

    topLevelTags := mb.identifyTopLevelTags()
    for _, tag := range topLevelTags {
        icon := mb.getIconForTag(tag)
        sb.WriteString(fmt.Sprintf("- %s [[%s MOC]]\n", icon, tag))
    }

    if mb.config.IncludeStats {
        sb.WriteString("\n## 📊 Quick Stats\n\n")
        sb.WriteString(fmt.Sprintf("- Total Notes: %d\n", len(mb.notes)))
        sb.WriteString(fmt.Sprintf("- Total Tags: %d\n", len(mb.tags)))
    }

    if mb.config.IncludeRecent {
        sb.WriteString("\n## 🔍 Recently Updated\n\n")
        recentNotes := mb.getRecentNotes(mb.config.RecentCount)
        for _, note := range recentNotes {
            sb.WriteString(fmt.Sprintf("- [[%s]]\n", note.Title))
        }
    }

    return MOCFile{
        Filename: "Home.md",
        Title:    "Home",
        Content:  sb.String(),
        Tags:     []string{"moc", "index"},
        Order:    0,
    }
}

func (mb *MOCBuilder) createTagMOC(tag string) MOCFile {
    var sb strings.Builder

    notes := mb.tags[tag]

    sb.WriteString("---\n")
    sb.WriteString(fmt.Sprintf("title: %s\n", strings.Title(tag)))
    sb.WriteString(fmt.Sprintf("tags: [moc, %s]\n", tag))
    sb.WriteString("---\n\n")
    sb.WriteString(fmt.Sprintf("# %s %s\n\n", mb.getIconForTag(tag), strings.Title(tag)))

    // Group notes by subtags if they exist
    subtags := mb.getSubtags(tag)
    if len(subtags) > 0 {
        for _, subtag := range subtags {
            subtagNotes := mb.getNotesWithTag(subtag)
            sb.WriteString(fmt.Sprintf("## %s\n\n", strings.Title(subtag)))
            for _, note := range subtagNotes {
                sb.WriteString(fmt.Sprintf("- [[%s]]\n", note.Title))
            }
            sb.WriteString("\n")
        }
    } else {
        // List all notes
        sb.WriteString("## Notes\n\n")
        for _, note := range notes {
            sb.WriteString(fmt.Sprintf("- [[%s]]\n", note.Title))
        }
    }

    sb.WriteString(fmt.Sprintf("\n---\n**Tagged Notes**: #%s (%d notes)\n", tag, len(notes)))

    return MOCFile{
        Filename: fmt.Sprintf("%s MOC.md", strings.Title(tag)),
        Title:    fmt.Sprintf("%s MOC", strings.Title(tag)),
        Content:  sb.String(),
        Tags:     []string{"moc", tag},
        Order:    1,
    }
}

func (mb *MOCBuilder) identifyTopLevelTags() []string {
    // Identify top-level tags based on:
    // 1. Frequency (>10% of notes)
    // 2. Broad categories (work, personal, learning, etc.)
    // 3. Tags without parent hierarchy

    tagScores := make(map[string]float64)
    totalNotes := float64(len(mb.notes))

    for tag, notes := range mb.tags {
        // Score based on frequency
        frequency := float64(len(notes)) / totalNotes
        score := frequency

        // Boost score for known top-level categories
        topLevelCategories := []string{"work", "personal", "learning", "projects", "ideas", "reference"}
        for _, cat := range topLevelCategories {
            if strings.EqualFold(tag, cat) {
                score += 0.5
            }
        }

        // Reduce score for tags with hierarchy
        if strings.Contains(tag, "/") || strings.Contains(tag, ".") {
            score *= 0.3
        }

        tagScores[tag] = score
    }

    // Sort tags by score and take top ones
    type tagScore struct {
        tag   string
        score float64
    }

    var sorted []tagScore
    for tag, score := range tagScores {
        sorted = append(sorted, tagScore{tag, score})
    }

    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].score > sorted[j].score
    })

    // Take top 5-10 tags as top-level
    maxTopLevel := 10
    if len(sorted) < maxTopLevel {
        maxTopLevel = len(sorted)
    }

    result := make([]string, maxTopLevel)
    for i := 0; i < maxTopLevel; i++ {
        result[i] = sorted[i].tag
    }

    return result
}

func (mb *MOCBuilder) getIconForTag(tag string) string {
    icons := map[string]string{
        "work":     "💼",
        "personal": "🏠",
        "learning": "📚",
        "projects": "🚀",
        "ideas":    "💡",
        "reference":"📖",
        "meetings": "🤝",
        "planning": "📋",
        "security": "🔐",
        "code":     "💻",
        "design":   "🎨",
    }

    if icon, exists := icons[strings.ToLower(tag)]; exists {
        return icon
    }

    return "📄"
}
```

## CLI Integration

### Command Handler (`cmd/sncli/migrate.go`)

```go
package main

import (
    "fmt"
    "github.com/urfave/cli/v2"
    "github.com/jonhadfield/gosn-v2/cache"
    "github.com/jonhadfield/gosn-v2/common"
    sncli "github.com/jonhadfield/sn-cli/internal/sncli"
)

func cmdMigrate() *cli.Command {
    return &cli.Command{
        Name:  "migrate",
        Usage: "migrate notes to other applications",
        Description: `Export your Standard Notes to other note-taking applications
with intelligent organization and automatic MOC generation.

Supported providers:
  - obsidian: Export to Obsidian vault (markdown + wikilinks)

Example:
  sn migrate obsidian --output ./my-vault --moc`,
        Subcommands: []*cli.Command{
            {
                Name:    "obsidian",
                Aliases: []string{"obs"},
                Usage:   "migrate to Obsidian vault",
                Flags: []cli.Flag{
                    &cli.StringFlag{
                        Name:     "output",
                        Aliases:  []string{"o"},
                        Usage:    "output directory for Obsidian vault",
                        Required: true,
                    },
                    &cli.BoolFlag{
                        Name:    "moc",
                        Aliases: []string{"m"},
                        Usage:   "generate Maps of Content (MOCs)",
                        Value:   true,
                    },
                    &cli.StringFlag{
                        Name:    "moc-style",
                        Usage:   "MOC generation style: flat, hierarchical, para, topic, auto",
                        Value:   "flat",
                    },
                    &cli.IntFlag{
                        Name:  "moc-depth",
                        Usage: "maximum MOC hierarchy depth",
                        Value: 2,
                    },
                    &cli.StringFlag{
                        Name:  "tag-filter",
                        Usage: "only export notes with these tags (comma-separated)",
                    },
                    &cli.BoolFlag{
                        Name:  "preserve-uuid",
                        Usage: "include Standard Notes UUID in frontmatter",
                        Value: true,
                    },
                    &cli.StringFlag{
                        Name:  "link-style",
                        Usage: "link style: wikilink, markdown, relative",
                        Value: "wikilink",
                    },
                    &cli.StringFlag{
                        Name:  "tag-style",
                        Usage: "tag style: frontmatter, inline, both",
                        Value: "frontmatter",
                    },
                    &cli.BoolFlag{
                        Name:  "dry-run",
                        Usage: "preview migration without writing files",
                    },
                },
                Action: func(c *cli.Context) error {
                    opts := getOpts(c)
                    return processMigrateObsidian(c, opts)
                },
            },
        },
    }
}

func processMigrateObsidian(c *cli.Context, opts configOptsOutput) error {
    // Get session
    session, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
    if err != nil {
        return err
    }

    var cacheDBPath string
    cacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
    if err != nil {
        return err
    }
    session.CacheDBPath = cacheDBPath

    // Parse tag filter
    var tagFilter []string
    if tagStr := c.String("tag-filter"); tagStr != "" {
        tagFilter = sncli.CommaSplit(tagStr)
    }

    // Create migration config
    migrateConfig := sncli.MigrateConfig{
        Session:      &session,
        Provider:     "obsidian",
        OutputDir:    c.String("output"),
        GenerateMOCs: c.Bool("moc"),
        MOCStyle:     sncli.MOCStyle(c.String("moc-style")),
        MOCDepth:     c.Int("moc-depth"),
        TagFilter:    tagFilter,
        DryRun:       c.Bool("dry-run"),
        Debug:        opts.debug,
    }

    // Execute migration
    result, err := migrateConfig.Run()
    if err != nil {
        return fmt.Errorf("migration failed: %w", err)
    }

    // Display results
    displayMigrationResults(c, result)

    return nil
}

func displayMigrationResults(c *cli.Context, result *sncli.MigrationResult) {
    fmt.Fprintf(c.App.Writer, "\n✅ Migration completed successfully!\n\n")
    fmt.Fprintf(c.App.Writer, "📊 Summary:\n")
    fmt.Fprintf(c.App.Writer, "  - Notes exported: %d\n", result.NotesExported)
    fmt.Fprintf(c.App.Writer, "  - MOCs created: %d\n", result.MOCsCreated)
    fmt.Fprintf(c.App.Writer, "  - Tags processed: %d\n", result.TagsProcessed)
    fmt.Fprintf(c.App.Writer, "  - Duration: %s\n", result.Duration)
    fmt.Fprintf(c.App.Writer, "  - Output: %s\n", result.OutputPath)

    if len(result.Warnings) > 0 {
        fmt.Fprintf(c.App.Writer, "\n⚠️  Warnings:\n")
        for _, warning := range result.Warnings {
            fmt.Fprintf(c.App.Writer, "  - %s\n", warning)
        }
    }

    if len(result.Errors) > 0 {
        fmt.Fprintf(c.App.Writer, "\n❌ Errors:\n")
        for _, err := range result.Errors {
            fmt.Fprintf(c.App.Writer, "  - %s\n", err)
        }
    }

    fmt.Fprintf(c.App.Writer, "\n🎉 Your Obsidian vault is ready at: %s\n", result.OutputPath)
    fmt.Fprintf(c.App.Writer, "💡 Open it in Obsidian to start exploring!\n\n")
}
```

## Advanced Features

### 1. Smart MOC Generation with Content Analysis

```go
type ContentAnalyzer struct {
    notes []Note
}

func (ca *ContentAnalyzer) AnalyzeTopics() map[string][]Note {
    // Use TF-IDF to identify important terms
    // Cluster notes by content similarity
    // Generate topic-based MOCs
}

func (ca *ContentAnalyzer) ExtractKeywords(note Note) []string {
    // NLP-based keyword extraction
    // Could use external library like prose or similar
}

func (ca *ContentAnalyzer) FindRelatedNotes(note Note) []Note {
    // Cosine similarity between notes
    // Return most related notes for "See Also" sections
}
```

### 2. Link Conversion

```go
type LinkConverter struct {
    notesByUUID  map[string]Note
    notesByTitle map[string]Note
}

func (lc *LinkConverter) ConvertLinks(content string) string {
    // Find Standard Notes references/links
    // Convert to Obsidian wikilinks
    // Handle broken links gracefully
}
```

### 3. Template System

```go
type MOCTemplate struct {
    Name        string
    Description string
    Template    string
    Variables   map[string]string
}

var PredefinedTemplates = map[string]MOCTemplate{
    "zettelkasten": {
        Name: "Zettelkasten",
        Description: "Classic Zettelkasten structure",
        // Template content
    },
    "para": {
        Name: "PARA",
        Description: "Projects, Areas, Resources, Archives",
        // Template content
    },
    "academic": {
        Name: "Academic Research",
        Description: "Academic research organization",
        // Template content
    },
}
```

## Testing Strategy

```go
// internal/sncli/migrate_test.go

func TestMigrateObsidian(t *testing.T) {
    // Test basic export
}

func TestMOCGeneration(t *testing.T) {
    // Test MOC generation with various styles
}

func TestLinkConversion(t *testing.T) {
    // Test link conversion from SN to Obsidian format
}

func TestTagHierarchy(t *testing.T) {
    // Test hierarchical tag detection and MOC generation
}

func TestFilenameS sanitization(t *testing.T) {
    // Test filename sanitization
}
```

## Documentation Updates

Add to README.md:

```markdown
### 📤 Migration

Export your notes to other applications:

```bash
# Basic Obsidian export
sn migrate obsidian --output ./my-vault

# With MOC generation
sn migrate obsidian --output ./vault --moc --moc-style hierarchical

# Filter by tags
sn migrate obsidian --output ./vault --tag-filter work,projects

# Preview without writing
sn migrate obsidian --output ./vault --dry-run
```

**MOC Styles:**
- `flat`: Single-level MOCs (recommended for most users)
- `hierarchical`: Nested MOC structure based on tag hierarchy
- `para`: Projects, Areas, Resources, Archives method
- `topic`: Topic/domain-based organization
- `auto`: AI-powered intelligent categorization
```

## Implementation Timeline

### Phase 1: Basic Export (Week 1)
- Core migration infrastructure
- Basic Obsidian export
- Simple file conversion
- Filename sanitization

### Phase 2: MOC Generation (Week 2)
- Tag analysis
- Flat MOC generation
- Home MOC creation
- Category MOCs

### Phase 3: Advanced MOCs (Week 3)
- Hierarchical MOCs
- PARA method
- Topic-based MOCs
- Template system

### Phase 4: Polish & Testing (Week 4)
- Comprehensive testing
- Documentation
- Error handling
- Performance optimization
- User feedback integration

## Future Enhancements

1. **Other Providers**
   - Logseq export
   - Notion API integration
   - Joplin export
   - Bear export

2. **Advanced Features**
   - Bidirectional sync
   - Incremental updates
   - Attachment handling
   - Rich content preservation

3. **AI Features**
   - Automatic topic detection
   - Smart categorization
   - Relationship discovery
   - MOC optimization suggestions

4. **Customization**
   - Custom templates
   - MOC styling options
   - Filename patterns
   - Frontmatter customization
