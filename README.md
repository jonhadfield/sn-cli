# sn-cli

> Manage your [Standard Notes](https://standardnotes.com/) account from the terminal.

<!-- Badges belong on one line so they render side by side. -->
<!-- markdownlint-disable MD013 -->

[![Tests](https://github.com/jonhadfield/sn-cli/actions/workflows/tests.yml/badge.svg)](https://github.com/jonhadfield/sn-cli/actions/workflows/tests.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/jonhadfield/sn-cli)](https://goreportcard.com/report/github.com/jonhadfield/sn-cli) [![Latest release](https://img.shields.io/github/v/release/jonhadfield/sn-cli)](https://github.com/jonhadfield/sn-cli/releases/latest) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

<!-- markdownlint-enable MD013 -->

Standard Notes is an end-to-end encrypted notes app. `sn` is an unofficial
command-line client for it: it signs in, syncs, and decrypts your items
locally, so you can read, write, search and back up your notes without opening
the desktop or web app. It works against the hosted service and against a
self-hosted server.

It is a single static binary with no runtime dependencies.

<!-- TODO: a terminal output sample belongs here - ideally an asciinema cast or
     a couple of screenshots of `sn stats --visual`, `sn tag cloud` and
     `sn search -q ... --output rich`, which are the commands whose output is
     worth seeing before you install anything. -->

## Features

- Notes, tags, checklists and note templates, all from the shell
- Full-text search over titles and content, with fuzzy matching and an offline mode
- Backup and restore to a single zip file, optionally encrypted
- Export to Markdown, HTML or JSON, including Hugo and Jekyll layouts
- Migration to an Obsidian vault, with Maps of Content generated for you
- Sessions stored in the system keychain on macOS and Linux, so your password
  need not live in your environment
- Runs on macOS, Linux, Windows, FreeBSD and OpenBSD

## Installation

On macOS and Linux, using [homebrew](https://brew.sh):

```bash
brew install jonhadfield/tap/sn-cli
```

That installs the `sn` binary and clears the macOS quarantine flag for you.

**Or in one line**, which picks the right archive for your platform, verifies
its checksum against the published list, and installs to `/usr/local/bin`,
using sudo only if the install directory is not writable:

<!-- The point of these is that they are a single pasteable line, so they
     cannot be wrapped to the usual width. -->
<!-- markdownlint-disable MD013 -->

```bash
curl -fsSL https://raw.githubusercontent.com/jonhadfield/sn-cli/main/install.sh | sh
```

Set `BIN_DIR` to install somewhere else, or `VERSION` to pin a release. With a
`BIN_DIR` you own, such as the one below, it needs no sudo at all:

```bash
curl -fsSL https://raw.githubusercontent.com/jonhadfield/sn-cli/main/install.sh | BIN_DIR="$HOME/.local/bin" sh
curl -fsSL https://raw.githubusercontent.com/jonhadfield/sn-cli/main/install.sh | VERSION=0.5.3 sh
```

<!-- markdownlint-enable MD013 -->

If you would rather read it before running it, [install.sh](install.sh) is in
this repository.

Otherwise, download an archive for your platform from the
[releases page](https://github.com/jonhadfield/sn-cli/releases) and put `sn`
somewhere on your `PATH`. The darwin binaries are not signed, so a tarball
fetched through a browser is quarantined and Gatekeeper will refuse to run it.
Homebrew and `install.sh` handle this; if you downloaded it yourself, clear the
flag:

```bash
xattr -d com.apple.quarantine /usr/local/bin/sn
```

## Getting started

You need a Standard Notes account. If you do not have one, `sn register --email
you@example.com` will create it, but read [Known issues](#known-issues) first.

The recommended way to sign in is to store a session in your system keychain,
so your password is neither typed for every command nor left in your
environment:

```bash
# Prompts for your email, password and 2FA code if you have it enabled,
# then stores the session in the keychain (macOS and Linux).
sn session --add

# Tell sn to use it, either for the shell session or per command.
export SN_USE_SESSION=true
sn --use-session get notes
```

Once that is set up:

```bash
# What can it do?
sn --help

# Add a note
sn add note --title "My First Note" --text "Hello, Standard Notes!"

# List your notes
sn get notes

# Search them
sn search --query "hello"

# See what is in the account
sn stats
```

Check or remove the stored session at any time:

```bash
sn session --status
sn session --remove
```

To encrypt the stored session, give `--session-key` a value when you add it,
and the same value when you use it:

```bash
sn session --add --session-key "some-passphrase"
sn --session-key "some-passphrase" get notes
```

### Signing in without a stored session

For CI and scripting, `sn` reads credentials from the environment:

```bash
export SN_EMAIL="you@example.com"
export SN_PASSWORD="your-password"
```

Be deliberate about this. `SN_PASSWORD` puts your Standard Notes password —
the key to an end-to-end encrypted account — in plaintext in your environment,
and in your shell history if you export it interactively. Prefer `sn session
--add` for day-to-day use, and a secret manager for automation.

### Sessions on headless servers

`sn session --add` stores the session in the system keychain, which a headless
Linux server usually does not have. Point `sn` at a file instead, sign in once
over SSH — entering your 2FA code if you have it enabled — and later commands
refresh the session and write it back to that file:

```bash
export SN_SESSION_FILE=~/.config/sncli/session
export SN_USE_SESSION=true

sn session --add
sn get notes
```

The file is written with `0600` permissions, so only you can read it. It holds
a session rather than your password, and `--session-key` encrypts it just as it
does a keychain session. The path can also be given with `--session-file`, or
as `session_file` in the config file.

## Commands

Run `sn <command> --help` for the flags of any command.

| Command | Description |
|---------|-------------|
| `add` | Add notes or tags |
| `backup` | Create, inspect and restore zip backups (alias: `bak`) |
| `debug` | Debugging tools, such as decrypting a single string |
| `delete` | Delete notes, tags or items by title or UUID, or clear duplicates |
| `edit` | Edit a note or tag in your `$EDITOR` |
| `editor` | Show or set the editor Standard Notes associates with new notes |
| `export` | Export notes to Markdown, HTML or JSON (alias: `exp`) |
| `get` | Retrieve notes, tags, items or account settings |
| `healthcheck` | Find account data errors, such as items that will not decrypt |
| `migrate` | Migrate notes to another application (currently Obsidian) |
| `organize` | Suggest tags and better titles using Google Gemini |
| `register` | Register a new Standard Notes account |
| `resync` | Purge the local cache and sync again |
| `search` | Full-text search across notes (alias: `find`) |
| `session` | Add, inspect and remove the keychain session |
| `stats` | Show statistics about the account |
| `tag` | Apply tags in bulk, and visualise them |
| `task` | Manage Checklist and Advanced Checklist tasks |
| `template` | Manage note templates (alias: `tpl`) |
| `wipe` | Delete all supported content in the account |

These global flags work on every command:

| Flag | Description |
|------|-------------|
| `--debug` | Verbose output, useful when reporting a bug |
| `--server` | Use a self-hosted server rather than the hosted service |
| `--use-session` | Authenticate with the stored keychain session |
| `--session-key` | Passphrase for an encrypted stored session |
| `--cachedb-dir` | Where to keep the local cache (default `~/.sn-cli`) |

## Examples

```bash
# Create a note, with tags
sn add note --title "Meeting Notes" --text "Important discussion points" --tag work,meetings

# Create a note from a file
sn add note --file ./notes/standup.md --tag work

# Create a nested tag
sn add tag --title "projects" --parent work

# Find notes by tag, with a preview of each
sn get notes --tag work --preview

# Read a note with markdown rendering
sn get notes --title "Meeting Notes" --rich

# Edit a note in $EDITOR
sn edit note --title "Meeting Notes"

# See which notes the app marked as copies, then delete them
sn delete duplicates --dry-run
sn delete duplicates

# Tag every note whose title matches, in one go
sn tag --find-title "invoice" --title finance --ignore-case

# See which tags are actually being used
sn tag stats
sn tag cloud

# Add and complete a checklist task
sn task add --list "Todo" --title "Buy groceries"
sn task show --list "Todo"
sn task complete --list "Todo" --title "Buy groceries"

# Statistics, with charts
sn stats --visual
```

### Search

```bash
# Titles and content are both searched by default
sn search --query "keyword"
sn search -q "keyword"

# Fuzzy matching, for when you cannot remember the spelling
sn search -q "imprtant" --fuzzy

# Case-sensitive, limited to one tag
sn search -q "Project" --case-sensitive --tag work

# Top ten results, rendered with markdown highlighting
sn search -q "todo" --limit 10 --output rich

# Titles only, which is faster
sn search -q "meeting" --content=false

# Skip the sync and search the local cache only
sn search -q "meeting" --offline
```

Results are sorted by relevance, with title matches scoring above content
matches, and shown with the matching terms highlighted in context. `--output`
accepts `table` (the default), `rich`, `json` and `yaml`.

Run `sn search --help` for the full set of flags.

### Backup and restore

```bash
# Everything, into one zip
sn backup create --output ./sn-backup.zip

# Encrypted, and only what changed since the last one
sn backup create --output ./sn-backup.zip --encrypt --incremental

# Look inside a backup before trusting it
sn backup info --file ./sn-backup.zip

# Restore, with a dry run first
sn backup restore --input ./sn-backup.zip --dry-run
sn backup restore --input ./sn-backup.zip
```

### Export

```bash
# Markdown, one file per note
sn export --output ./export

# HTML, foldered by tag, with frontmatter
sn export --output ./export --format html --by-tags --metadata

# Ready to drop into a Hugo or Jekyll site
sn export --output ./site/content --static-site hugo
```

### Note templates

<!-- markdownlint-disable MD013 -->

```bash
sn template list
sn template create --name standup --title "Standup {{date}}" --content "## Yesterday" --tags work
sn template show --name standup
sn template use --template standup --title "Standup" --var project=sn-cli
```

<!-- markdownlint-enable MD013 -->

### Organising notes with AI

`sn organize` asks Google Gemini to suggest tags and better titles for your
notes, shows you a preview, and applies nothing until you agree.

**This command sends the content of the notes it processes to Google Gemini.**
No other part of `sn` transmits note content anywhere except to your Standard
Notes server. If that is not acceptable for your notes, do not use this
command, or narrow what it sees with `--title`, `--uuid`, `--since` and
`--until`.

```bash
# Needs an API key, from the flag or GEMINI_API_KEY
export GEMINI_API_KEY="..."

# Preview suggestions for recent notes and confirm interactively
sn organize --since 2026-01-01T00:00:00Z

# Only the notes whose titles match
sn organize --title "meeting,standup"

# Apply without confirming
sn organize --since 2026-01-01T00:00:00Z --yes
```

Keys come from [Google AI Studio](https://aistudio.google.com/app/apikey).

### Migrating to Obsidian

```bash
# An Obsidian vault, with Maps of Content generated for you
sn migrate obsidian --output ./my-vault

# See what it would write, without writing it
sn migrate obsidian --output ./my-vault --dry-run
```

Tags become YAML frontmatter, links become wikilinks, and creation dates and
UUIDs are preserved. See [docs/obsidian-migration.md](docs/obsidian-migration.md)
for the MOC styles, tag filtering and the vault layout it produces.

## Configuration

There is no config file. `sn` is configured with flags and environment
variables:

| Variable | Effect |
|----------|--------|
| `SN_EMAIL` | Account email, when not using a stored session |
| `SN_PASSWORD` | Account password, when not using a stored session |
| `SN_SERVER` | Self-hosted server URL |
| `SN_USE_SESSION` | Set to `true` to authenticate with the keychain session |
| `SN_CACHEDB_DIR` | Where to keep the local cache (default `~/.sn-cli`) |
| `SN_DEBUG` | Set to `true` for verbose output |
| `SN_DEFAULT_LIST` | Default `--list` for the `task` command |
| `SN_DEFAULT_GROUP` | Default `--group` for the `task` command |
| `SN_SHOW_COMPLETED` | Set to `true` to include completed tasks by default |
| `GEMINI_API_KEY` | API key for `sn organize` |
| `EDITOR` | Editor used by `sn edit note` |

`--session-key` is a flag only; it is deliberately not read from the
environment.

### Self-hosted servers

```bash
export SN_SERVER="https://your-standardnotes-server.com"
```

### Shell completion

Tab completion is available for Bash, Zsh, Fish and PowerShell. The scripts call
the binary's built-in `--generate-bash-completion` flag, so they stay in step
with the commands and flags automatically.

The completion script has to be installed under the name of the binary, which
is `sn`:

```bash
# Bash, on macOS with homebrew
brew install bash-completion@2
sudo cp autocomplete/bash_autocomplete "$(brew --prefix)/etc/bash_completion.d/sn"
```

See [autocomplete/README.md](autocomplete/README.md) for the other shells and
for troubleshooting.

## Troubleshooting

**Notes are missing, or out of date.** `sn` keeps a local cache under
`~/.sn-cli` (override with `--cachedb-dir` or `SN_CACHEDB_DIR`). Purge it and
sync again:

```bash
sn resync
```

**Items that will not decrypt.** Usually a sign that an items key is missing or
damaged, which can happen after a password change or a partial sync elsewhere:

```bash
sn healthcheck keys
```

**"sn" cannot be opened on macOS.** The darwin binaries are unsigned, so a
tarball downloaded through a browser carries the quarantine flag. Homebrew and
`install.sh` clear it; otherwise do it yourself:

```bash
xattr -d com.apple.quarantine /usr/local/bin/sn
```

**Tab completion does nothing.** The completion script has to be installed
under the binary's name, `sn`, not `sncli`. See
[autocomplete/README.md](autocomplete/README.md).

**Something fails and you cannot tell why.** Run the command again with
`--debug`, which reports the requests being made. Include that output if you
open an issue, with your password and session redacted.

## Known issues

- New accounts registered with `sn register` need one login through the
  official web or desktop app before `sn` can use them, because that first
  login is what initialises the account's encryption keys.

## Development

```bash
git clone https://github.com/jonhadfield/sn-cli.git
cd sn-cli

# Builds .local_dist/sncli
make build

# Build and install it to /usr/local/bin/sn
make mac-install     # or: make linux-install

# Run the offline unit tests
make test

# Lint and test, as CI does
make ci
```

Tests that talk to a real Standard Notes account are opt-in, because they
create and delete live items. `make test` and CI run only the offline unit
tests. To run the live ones as well:

```bash
SN_INTEGRATION_TESTS=1 SN_EMAIL=you@example.com SN_PASSWORD=... go test ./...
```

## Contributing

Contributions are welcome. Please read the [Contributing
Guide](CONTRIBUTING.md) for the code of conduct and the process for submitting
pull requests.

## License

MIT. See [LICENSE](LICENSE).

## Links

- [Standard Notes](https://standardnotes.com/) - the note-taking app this CLI supports
- [Changelog](CHANGELOG.md) - what changed, release by release
- [Releases](https://github.com/jonhadfield/sn-cli/releases) - download a build
- [Issues](https://github.com/jonhadfield/sn-cli/issues) - bugs and requests
