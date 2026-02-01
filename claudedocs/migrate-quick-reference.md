# Migration Feature Quick Reference

## Command Syntax

```bash
sn migrate <provider> --output <directory> [options]
```

## Providers

| Provider | Status | Format | Special Features |
|----------|--------|--------|------------------|
| `obsidian` | ✅ Planned | Markdown + Wikilinks | MOC generation, tags in frontmatter |
| `logseq` | 🔜 Future | Markdown + Outliner | Block references, namespaces |
| `notion` | 🔜 Future | API Upload | Databases, pages, relations |
| `joplin` | 🔜 Future | Markdown + Resources | Notebooks, attachments |

## Common Options

| Option | Short | Default | Description |
|--------|-------|---------|-------------|
| `--output` | `-o` | (required) | Output directory path |
| `--moc` | `-m` | `true` | Generate Maps of Content |
| `--moc-style` | - | `flat` | MOC organization style |
| `--moc-depth` | - | `2` | Maximum MOC hierarchy depth |
| `--tag-filter` | - | (all) | Only export specific tags |
| `--dry-run` | - | `false` | Preview without writing files |
| `--preserve-uuid` | - | `true` | Include SN UUID in frontmatter |
| `--link-style` | - | `wikilink` | Link format style |
| `--tag-style` | - | `frontmatter` | Tag placement style |

## MOC Styles

### `flat` (Default - Recommended)
```
Home.md
├─ Work MOC.md
├─ Learning MOC.md
├─ Personal MOC.md
└─ Projects MOC.md

All notes in root directory
```

**Best for**: Most users, small to medium vaults (<1000 notes)

### `hierarchical`
```
Home.md
Work/
├─ Work MOC.md
├─ Projects/
│  ├─ Projects MOC.md
│  ├─ Project Alpha.md
│  └─ Project Beta.md
└─ Meetings/
   ├─ Meetings MOC.md
   └─ ...notes...
```

**Best for**: Large collections, organized thinkers

### `para`
```
Home.md
1-Projects/
├─ Projects MOC.md
└─ ...active projects...
2-Areas/
├─ Areas MOC.md
└─ ...ongoing responsibilities...
3-Resources/
├─ Resources MOC.md
└─ ...reference material...
4-Archives/
└─ ...completed items...
```

**Best for**: GTD practitioners, productivity focus

### `topic`
```
Home.md
Security MOC.md
├─ Authentication MOC.md
├─ Web Security MOC.md
└─ Network Security MOC.md
Programming MOC.md
├─ Languages MOC.md
└─ Frameworks MOC.md
```

**Best for**: Academic research, deep learning

### `auto`
```
Uses AI/content analysis to determine:
- Optimal MOC structure
- Topic clustering
- Relationship mapping
- Smart categorization
```

**Best for**: Users wanting intelligent organization

## Quick Examples

### Basic Export
```bash
# Export all notes to Obsidian vault
sn migrate obsidian --output ./my-vault
```

### With MOC Generation
```bash
# Export with flat MOCs (recommended)
sn migrate obsidian --output ./vault --moc --moc-style flat

# Export with hierarchical MOCs
sn migrate obsidian --output ./vault --moc --moc-style hierarchical

# Export using PARA method
sn migrate obsidian --output ./vault --moc --moc-style para
```

### Filtered Export
```bash
# Only export work-related notes
sn migrate obsidian --output ./work-vault --tag-filter work

# Export multiple tag categories
sn migrate obsidian --output ./vault --tag-filter "work,projects,learning"

# Export with depth limit
sn migrate obsidian --output ./vault --moc --moc-depth 3
```

### Customization
```bash
# Use inline tags instead of frontmatter
sn migrate obsidian --output ./vault --tag-style inline

# Use markdown links instead of wikilinks
sn migrate obsidian --output ./vault --link-style markdown

# Don't preserve UUIDs
sn migrate obsidian --output ./vault --preserve-uuid=false
```

### Preview Mode
```bash
# Dry run to see what would be exported
sn migrate obsidian --output ./vault --dry-run

# Preview with detailed output
sn migrate obsidian --output ./vault --dry-run --debug
```

## Obsidian-Specific Options

### Link Styles

#### `wikilink` (Default)
```markdown
See [[Project Requirements]] for details.
Related: [[Meeting Notes]]
```

#### `markdown`
```markdown
See [Project Requirements](project-requirements.md) for details.
Related: [Meeting Notes](meeting-notes.md)
```

#### `relative`
```markdown
See [Project Requirements](../projects/project-requirements.md)
```

### Tag Styles

#### `frontmatter` (Default)
```markdown
---
tags: [work, projects, important]
---

# Project Requirements
```

#### `inline`
```markdown
# Project Requirements

#work #projects #important
```

#### `both`
```markdown
---
tags: [work, projects, important]
---

# Project Requirements

#work #projects #important
```

## Output Structure

### File Naming
```
Original: "Meeting Notes: Q1 Planning"
Sanitized: "Meeting Notes - Q1 Planning.md"

Characters removed: : " * ? < > |
Spaces preserved: Yes
Max length: 255 characters
Duplicates: Numbered (note.md, note-1.md, note-2.md)
```

### Frontmatter Template
```yaml
---
title: "Note Title"
tags: [tag1, tag2, tag3]
created: 2024-01-15T10:30:00Z
updated: 2024-01-15T14:20:00Z
uuid: abc123-def456-ghi789
source: standard-notes
aliases: []  # Optional
---
```

## Expected Output

### Small Vault (<100 notes)
```
📊 Migration Statistics:
  - Notes exported: 87
  - MOCs created: 6
  - Tags processed: 23
  - Duration: 3.2s
  - Output: ./my-vault

📁 Structure:
  - Home.md (entry point)
  - 5 category MOCs
  - 87 note files
  - Total files: 93
```

### Medium Vault (100-500 notes)
```
📊 Migration Statistics:
  - Notes exported: 342
  - MOCs created: 12
  - Tags processed: 45
  - Duration: 12.8s
  - Output: ./my-vault

📁 Structure:
  - Home.md (entry point)
  - 11 category MOCs
  - 342 note files
  - Total files: 354
```

### Large Vault (>500 notes)
```
📊 Migration Statistics:
  - Notes exported: 1,247
  - MOCs created: 28
  - Tags processed: 89
  - Duration: 45.3s
  - Output: ./my-vault

📁 Structure:
  - Home.md (entry point)
  - 27 category MOCs (2 levels deep)
  - 1,247 note files
  - Total files: 1,276
```

## Troubleshooting

### Common Issues

#### Output directory already exists
```bash
Error: output directory already exists: ./my-vault

Solution: Use different directory or remove existing one
```

#### No notes found
```bash
Error: no notes found to export

Solution: Run 'sn resync' first to refresh cache
```

#### Permission denied
```bash
Error: permission denied: ./my-vault

Solution: Check directory permissions or use different location
```

#### Tag filter no matches
```bash
Warning: tag filter returned 0 notes

Solution: Check tag names (case-sensitive) or remove filter
```

## Post-Migration Steps

1. **Open in Obsidian**
   ```bash
   # Open Obsidian and select the output directory as vault
   ```

2. **Review Home MOC**
   ```
   Start at Home.md to navigate your knowledge base
   ```

3. **Check MOC Structure**
   ```
   Verify MOCs match your organization preferences
   ```

4. **Customize**
   ```
   - Adjust MOC templates
   - Reorganize if needed
   - Add custom CSS
   - Configure plugins
   ```

5. **Backup**
   ```bash
   # Backup your new vault
   zip -r vault-backup.zip ./my-vault
   ```

## Performance Tips

### For Large Vaults
```bash
# Use tag filtering to split migration
sn migrate obsidian --output ./work --tag-filter work
sn migrate obsidian --output ./personal --tag-filter personal

# Disable MOC generation for speed
sn migrate obsidian --output ./vault --moc=false

# Use simpler MOC style
sn migrate obsidian --output ./vault --moc-style flat
```

### For Best MOC Quality
```bash
# Use auto style with AI analysis
sn migrate obsidian --output ./vault --moc-style auto

# Increase MOC depth for detailed organization
sn migrate obsidian --output ./vault --moc-depth 3
```

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Standard Notes                            │
│                   (Source Data)                              │
└────────────────────┬────────────────────────────────────────┘
                     │
                     │ Sync & Cache
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                   sn-cli Migration                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Analyze   │→ │  Transform  │→ │   Generate  │        │
│  │   Content   │  │   Format    │  │    MOCs     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└────────────────────┬────────────────────────────────────────┘
                     │
                     │ Write Files
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                  Obsidian Vault                              │
│                 (Output Files)                               │
│                                                              │
│  Home.md (Entry Point)                                      │
│  ├── Work MOC.md ─────────┐                                │
│  ├── Learning MOC.md ─────┤                                │
│  └── Personal MOC.md ─────┤                                │
│                            │                                 │
│  All Notes (Markdown)      │                                │
│  ├── note-1.md ◄───────────┘                               │
│  ├── note-2.md                                              │
│  └── note-n.md                                              │
└─────────────────────────────────────────────────────────────┘
```

## Related Documentation

- [Full Migration Plan](./migrate-feature-plan.md)
- [MOC Examples](./migrate-moc-examples.md)
- [Obsidian Documentation](https://help.obsidian.md/)
- [PARA Method](https://fortelabs.co/blog/para/)
- [Zettelkasten Method](https://zettelkasten.de/)
