# 📝 sn-cli

> A modern command-line interface for [Standard Notes](https://standardnotes.org/)

[![Build Status](https://www.travis-ci.org/jonhadfield/sn-cli.svg?branch=master)](https://www.travis-ci.org/jonhadfield/sn-cli) [![Go Report Card](https://goreportcard.com/badge/github.com/jonhadfield/sn-cli)](https://goreportcard.com/report/github.com/jonhadfield/sn-cli) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## ✨ Features

- **📋 Notes & Tasks**: Create, edit, and manage notes and checklists
- **🔍 Full-Text Search**: Search across titles and content with fuzzy matching and regex support
- **📤 Migration**: Export to Obsidian with automatic Maps of Content (MOC) generation
- **🏷️ Tags**: Organize content with flexible tagging
- **📊 Statistics**: Detailed analytics about your notes and usage
- **🔐 Secure Sessions**: Keychain integration for macOS and Linux
- **⚡ Fast Sync**: Efficient synchronization with Standard Notes servers
- **🔄 Multi-Platform**: Windows, macOS, and Linux support

## 🚀 Quick Start

### Installation

**Download the latest release:**
```bash
# macOS/Linux
curl -L https://github.com/jonhadfield/sn-cli/releases/latest/download/sncli_$(uname -s)_$(uname -m) -o sn
chmod +x sn && sudo mv sn /usr/local/bin/

# Or via direct download
# Visit: https://github.com/jonhadfield/sn-cli/releases
```

### First Run

```bash
# See all available commands
sn --help

# Add a note
sn add note --title "My First Note" --text "Hello, Standard Notes!"

# List your notes
sn get notes

# View statistics
sn stats
```

## 📋 Commands

| Command | Description |
|---------|-------------|
| `add` | Add notes, tags, or tasks |
| `delete` | Delete items by title or UUID |
| `edit` | Edit existing notes |
| `get` | Retrieve notes, tags, or tasks |
| `search` | Full-text search across notes (supports fuzzy matching and regex) |
| `migrate` | Migrate notes to other applications (Obsidian, etc.) with MOC generation |
| `tag` | Manage tags and tagging |
| `task` | Manage checklists and advanced checklists |
| `stats` | Display detailed statistics |
| `session` | Manage stored sessions |
| `register` | Register a new Standard Notes account |
| `resync` | Refresh local cache |
| `wipe` | Delete all notes and tags |

*Note: Export and import are temporarily disabled due to recent Standard Notes API changes*

## 🔐 Authentication

### Environment Variables
```bash
export SN_EMAIL="your-email@example.com"
export SN_PASSWORD="your-password"
export SN_SERVER="https://api.standardnotes.com"  # Optional for self-hosted
```

### Session Storage (Recommended)
Store encrypted sessions in your system keychain:

```bash
# Add session (supports 2FA)
sn session --add

# Add encrypted session
sn session --add --session-key

# Use session automatically
export SN_USE_SESSION=true
# or
sn --use-session get notes
```

## 🆕 Recent Updates

### Version 0.4.1 (2026-01-30)
- 🔐 **Fixed**: Authentication issues with updated dependencies
- 🏷️ **Improved**: Tag cloud visualization with offline support
- 🛡️ **Enhanced**: Network error handling and graceful degradation
- 🐛 **Fixed**: Tag reference matching and display issues

### Version 0.4.0 (2026-01-29)
- 💾 **Added**: Backup and restore functionality
- 📤 **Added**: Enhanced export with multiple formats
- 🎨 **Added**: Tag cloud visualization
- 📝 **Added**: Note templates system

**[View full changelog →](CHANGELOG.md)**

## 💡 Examples

```bash
# Create a note with tags
sn add note --title "Meeting Notes" --text "Important discussion points" --tag work,meetings

# Find notes by tag
sn get notes --tag work

# Search for notes (searches both title and content)
sn search --query "meeting"

# Fuzzy search with limit
sn search --query "mtng" --fuzzy --limit 5

# Case-sensitive regex search
sn search --query "TODO|FIXME" --regex --case-sensitive

# Search within specific tags
sn search --query "project" --tag work

# Create a checklist
sn add task --title "Todo List" --text "- Buy groceries\n- Call dentist\n- Finish project"

# View your note statistics
sn stats

# Edit a note
sn edit note --title "Meeting Notes" --text "Updated content"
```

### 🔍 Search Feature

The `search` command provides powerful full-text search across all your notes:

**Basic Usage:**
```bash
# Simple search
sn search --query "keyword"
sn search -q "keyword"  # Short form
```

**Search Options:**
```bash
--query, -q     Search query (required)
--content, -c   Search in note content (default: true)
--fuzzy, -f     Enable fuzzy matching for typo tolerance
--case-sensitive  Make search case-sensitive (default: false)
--tag           Filter results by tag
--limit, -l     Limit number of results (default: unlimited)
--output        Output format: table, rich, json, yaml (default: table)
```

**Advanced Examples:**
```bash
# Regex pattern matching
sn search -q "bug-[0-9]+" --regex

# Fuzzy search (matches similar terms)
sn search -q "imprtant" --fuzzy

# Case-sensitive search in work-tagged notes
sn search -q "Project" --case-sensitive --tag work

# Get top 10 results in rich format
sn search -q "todo" --limit 10 --output rich

# Search only in titles (faster)
sn search -q "meeting" --content=false
```

**Search Features:**
- Searches both note titles and content by default
- Highlights matching terms in results
- Shows context snippets around matches
- Sorts results by relevance (title matches score higher)
- Supports multiple output formats with syntax highlighting

### 📤 Migration to Other Applications

Export your notes to other platforms with intelligent organization:

```bash
# Basic export to Obsidian
sn migrate obsidian --output ./my-vault

# Export with automatic MOC generation
sn migrate obsidian --output ./vault --moc

# Export specific tags only
sn migrate obsidian --output ./vault --tag-filter work,projects

# Preview migration without writing files
sn migrate obsidian --output ./vault --dry-run
```

**Features:**
- Automatic MOC (Maps of Content) generation
- Tag preservation in YAML frontmatter
- Metadata preservation (dates, UUIDs)
- Multiple organizational styles
- Wikilink formatting

**Output Structure:**
```
my-vault/
├── Home.md              # Main entry point
├── Work MOC.md          # Category MOCs
├── Learning MOC.md
└── ... (all your notes)
```

## ⚙️ Advanced Configuration

### Shell Completion

Tab completion is available for Bash, Zsh, Fish, and PowerShell.

**Quick Install (Bash on macOS):**
```bash
brew install bash-completion@2
sudo cp autocomplete/bash_autocomplete /usr/local/etc/bash_completion.d/sncli
echo '[ -f /usr/local/etc/bash_completion ] && . /usr/local/etc/bash_completion' >> ~/.bash_profile
source ~/.bash_profile
```

**Quick Install (Fish):**
```bash
mkdir -p ~/.config/fish/completions
cp autocomplete/fish_autocomplete.fish ~/.config/fish/completions/sncli.fish
```

📖 **For detailed installation instructions for all shells, see [autocomplete/README.md](autocomplete/README.md)**

### Self-Hosted Servers
```bash
export SN_SERVER="https://your-standardnotes-server.com"
```

## 🔧 Development

```bash
# Build from source
git clone https://github.com/jonhadfield/sn-cli.git
cd sn-cli
make build

# Run tests
make test

# View all make targets
make help
```

## ⚠️ Known Issues

- New accounts registered via sn-cli require initial login through the official web/desktop app to initialize encryption keys

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- [Standard Notes](https://standardnotes.org/) - The note-taking app this CLI supports
- [Releases](https://github.com/jonhadfield/sn-cli/releases) - Download the latest version
- [Issues](https://github.com/jonhadfield/sn-cli/issues) - Report bugs or request features

---

