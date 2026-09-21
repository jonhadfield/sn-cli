# Shell Completion for sn

This directory contains shell completion scripts for the Standard Notes CLI.
The binary is called `sn`, and that name matters: the Bash and PowerShell
scripts take the command they complete from their own filename, so a script
installed under any other name completes a command that does not exist.

## Overview

The completion scripts use the CLI's built-in `--generate-bash-completion` flag
to generate completions dynamically, which means they stay in step with the
available commands and options automatically.

## Installation

### Bash

#### macOS with Homebrew

<!-- markdownlint-disable MD013 -->

```bash
# Install bash-completion if not already installed
brew install bash-completion@2

# Copy the completion script, named for the binary
sudo cp bash_autocomplete "$(brew --prefix)/etc/bash_completion.d/sn"

# Add to your ~/.bash_profile or ~/.bashrc
echo '[ -f "$(brew --prefix)/etc/bash_completion" ] && . "$(brew --prefix)/etc/bash_completion"' >> ~/.bash_profile
source ~/.bash_profile
```

<!-- markdownlint-enable MD013 -->

`$(brew --prefix)` is `/usr/local` on Intel Macs and `/opt/homebrew` on Apple
Silicon.

#### Linux

```bash
# Copy to system completion directory
sudo cp bash_autocomplete /etc/bash_completion.d/sn

# Source in your ~/.bashrc (usually automatic on next login)
echo 'source /etc/bash_completion.d/sn' >> ~/.bashrc
source ~/.bashrc
```

#### Manual Setup

```bash
# Add to your ~/.bashrc or ~/.bash_profile
export PROG=sn
source /path/to/sn-cli/autocomplete/bash_autocomplete
```

### Zsh

The zsh script reads the command name from `$PROG`, so set it before sourcing
the script:

```bash
# Add to your ~/.zshrc
autoload -U compinit && compinit
export PROG=sn
source /path/to/sn-cli/autocomplete/zsh_autocomplete

# Reload your shell
source ~/.zshrc
```

### PowerShell

Like the Bash script, this one takes the command name from its filename, so
save it as `sn.ps1`:

```powershell
# Find your profile location with: $PROFILE

# Copy the script to a permanent location, named for the binary
Copy-Item powershell_autocomplete.ps1 ~\Documents\WindowsPowerShell\sn.ps1

# Add to your profile
Add-Content $PROFILE ". ~\Documents\WindowsPowerShell\sn.ps1"

# Reload profile
. $PROFILE
```

### Fish

Fish loads completions from its completions directory automatically, so the
file only needs to be named after the binary:

```fish
mkdir -p ~/.config/fish/completions
cp fish_autocomplete.fish ~/.config/fish/completions/sn.fish
```

Open a new shell, or `source ~/.config/fish/completions/sn.fish` in the current
one, and `sn <TAB>` completes.

## Usage

Once installed, you can use Tab completion:

```bash
sn <TAB>              # Shows all commands
sn add <TAB>          # Shows add subcommands (note, tag)
sn get --<TAB>        # Shows available flags
```

## Verification

Test if completions are working:

```bash
# Type this and press TAB
sn a<TAB>

# Should show: add
```

## Troubleshooting

### Bash: "command not found: _get_comp_words_by_ref"

Install the bash-completion package:

- **macOS**: `brew install bash-completion@2`
- **Ubuntu/Debian**: `sudo apt-get install bash-completion`
- **Fedora/RHEL**: `sudo dnf install bash-completion`

### Completions not appearing

1. Check the script is installed under the name `sn`, not `sncli`. The Bash and
   PowerShell scripts derive the command they complete from their own filename.
2. Reload your shell: `exec $SHELL`, or open a new terminal.
3. Check that `sn` is on your PATH: `which sn`.

## References

- [urfave/cli Bash Completions](https://cli.urfave.org/v2/examples/bash-completions/)
- [Bash Completion Guide](https://github.com/scop/bash-completion)
- [Fish Shell Completions](https://fishshell.com/docs/current/completions.html)
- [PowerShell Completions](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/register-argumentcompleter)
