# 📝 sn-cli

> A modern command-line interface for [Standard Notes](https://standardnotes.org/)

[![Build Status](https://www.travis-ci.org/jonhadfield/sn-cli.svg?branch=master)](https://www.travis-ci.org/jonhadfield/sn-cli) [![Go Report Card](https://goreportcard.com/badge/github.com/jonhadfield/sn-cli)](https://goreportcard.com/report/github.com/jonhadfield/sn-cli) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## ✨ Features

- **📋 Notes & Tasks**: Create, edit, and manage notes and checklists
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

### Version 0.3.5 (2024-01-08)
- 🐛 **Fixed**: Conflict warning handling
- ✅ **Added**: Helper tests
- 🔧 **Improved**: Code simplification

### Version 0.3.4 (2024-01-07)
- 🐛 **Fixed**: Command completion and updated instructions

**[View full changelog →](CHANGELOG.md)**

## 💡 Examples

```bash
# Create a note with tags
sn add note --title "Meeting Notes" --text "Important discussion points" --tag work,meetings

# Find notes by tag
sn get notes --tag work

# Create a checklist
sn add task --title "Todo List" --text "- Buy groceries\n- Call dentist\n- Finish project"

# View your note statistics
sn stats

# Edit a note
sn edit note --title "Meeting Notes" --text "Updated content"
```

## ⚙️ Advanced Configuration

### Bash Completion

#### macOS (Homebrew)
```bash
brew install bash-completion
echo '[ -f /usr/local/etc/bash_completion ] && . /usr/local/etc/bash_completion' >> ~/.bash_profile
```

#### Install completion script
```bash
# macOS
cp bash_autocomplete /usr/local/etc/bash_completion.d/sn
echo "source /usr/local/etc/bash_completion.d/sn" >> ~/.bashrc

# Linux
cp bash_autocomplete /etc/bash_completion.d/sn
echo "source /etc/bash_completion.d/sn" >> ~/.bashrc
```

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

