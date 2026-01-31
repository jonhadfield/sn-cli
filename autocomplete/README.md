# Shell Completion for sncli

This directory contains shell completion scripts for the Standard Notes CLI (`sncli`).

## Overview

The completion scripts use the CLI's built-in `--generate-bash-completion` flag to dynamically generate completions, which means they automatically stay in sync with all available commands and options.

## Installation

### Bash

#### macOS with Homebrew

```bash
# Install bash-completion if not already installed
brew install bash-completion@2

# Copy the completion script
sudo cp bash_autocomplete /usr/local/etc/bash_completion.d/sncli

# Add to your ~/.bash_profile or ~/.bashrc
echo '[ -f /usr/local/etc/bash_completion ] && . /usr/local/etc/bash_completion' >> ~/.bash_profile
source ~/.bash_profile
```

#### Linux

```bash
# Copy to system completion directory
sudo cp bash_autocomplete /etc/bash_completion.d/sncli

# Source in your ~/.bashrc (usually automatic on next login)
echo 'source /etc/bash_completion.d/sncli' >> ~/.bashrc
source ~/.bashrc
```

#### Manual Setup

```bash
# Add to your ~/.bashrc or ~/.bash_profile
export PROG=sncli
source /path/to/sn-cli/autocomplete/bash_autocomplete
```

### Zsh

```bash
# Create completion directory if it doesn't exist
mkdir -p ~/.zsh/completion

# Copy the completion script
cp zsh_autocomplete ~/.zsh/completion/_sncli

# Add to your ~/.zshrc (if not already present)
fpath=(~/.zsh/completion $fpath)
autoload -U compinit && compinit

# Reload your shell
source ~/.zshrc
```

### Fish

```bash
# Fish automatically loads completions from this directory
mkdir -p ~/.config/fish/completions

# Copy the completion script
cp fish_autocomplete.fish ~/.config/fish/completions/sncli.fish

# Reload completions (or restart fish)
fish_update_completions
```

### PowerShell

```powershell
# Add to your PowerShell profile
# Find profile location with: $PROFILE

# Copy the script to a permanent location
Copy-Item powershell_autocomplete.ps1 ~\Documents\WindowsPowerShell\

# Add to your profile
Add-Content $PROFILE ". ~\Documents\WindowsPowerShell\powershell_autocomplete.ps1"

# Reload profile
. $PROFILE
```

## Usage

Once installed, you can use Tab completion:

```bash
sncli <TAB>              # Shows all commands
sncli add <TAB>          # Shows add subcommands (note, tag, task)
sncli get --<TAB>        # Shows available flags
```

## Verification

Test if completions are working:

```bash
# Type this and press TAB
sncli a<TAB>

# Should show: add
```

## Troubleshooting

### Bash: "command not found: _get_comp_words_by_ref"

Install the bash-completion package:
- **macOS**: `brew install bash-completion@2`
- **Ubuntu/Debian**: `sudo apt-get install bash-completion`
- **Fedora/RHEL**: `sudo dnf install bash-completion`

### Completions not appearing

1. Make sure the completion script is in the correct location
2. Reload your shell: `exec $SHELL` or open a new terminal
3. Check that `sncli` is in your PATH: `which sncli`

### Fish shell not finding completions

Make sure the file is named correctly: `~/.config/fish/completions/sncli.fish`

## References

- [urfave/cli Bash Completions](https://cli.urfave.org/v2/examples/bash-completions/)
- [Bash Completion Guide](https://github.com/scop/bash-completion)
- [Fish Shell Completions](https://fishshell.com/docs/current/completions.html)
- [PowerShell Completions](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/register-argumentcompleter)
