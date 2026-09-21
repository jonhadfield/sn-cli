# Fish shell completion for sn (Standard Notes CLI)
# Save this file to ~/.config/fish/completions/sn.fish

function __fish_sn_no_subcommand
    set -l cmd (commandline -opc)
    if [ (count $cmd) -eq 1 ]
        return 0
    end
    return 1
end

# Ask the CLI itself what it accepts here. The first token is the command as
# the user invoked it, so this keeps working if the binary is renamed.
function __fish_sn_complete
    set -l cmd (commandline -opc)

    if test (count $cmd) -gt 1
        # Get subcommand completions
        command $cmd[1] $cmd[2..-1] --generate-bash-completion 2>/dev/null
    else
        # Get top-level command completions
        command $cmd[1] --generate-bash-completion 2>/dev/null
    end
end

# Main completions
complete -c sn -f -n __fish_sn_no_subcommand -a '(__fish_sn_complete)' -d 'Standard Notes CLI'

# Global options
complete -c sn -l cachedb-dir -d 'Cache database directory'
complete -c sn -l debug -d 'Enable debug mode'
complete -c sn -l server -d 'Standard Notes server URL'
complete -c sn -l session-key -d 'Session encryption key'
complete -c sn -l use-session -d 'Use stored session'
complete -c sn -s h -l help -d 'Show help'
complete -c sn -s v -l version -d 'Show version'

# Command-specific completions
complete -c sn -n '__fish_seen_subcommand_from add' -a '(__fish_sn_complete)' -d 'Add items'
complete -c sn -n '__fish_seen_subcommand_from backup bak' -a '(__fish_sn_complete)' -d 'Backup operations'
complete -c sn -n '__fish_seen_subcommand_from delete' -a '(__fish_sn_complete)' -d 'Delete items'
complete -c sn -n '__fish_seen_subcommand_from edit' -a '(__fish_sn_complete)' -d 'Edit items'
complete -c sn -n '__fish_seen_subcommand_from editor' -a '(__fish_sn_complete)' -d 'Manage the editor for new notes'
complete -c sn -n '__fish_seen_subcommand_from export exp' -a '(__fish_sn_complete)' -d 'Export notes'
complete -c sn -n '__fish_seen_subcommand_from get' -a '(__fish_sn_complete)' -d 'Get items'
complete -c sn -n '__fish_seen_subcommand_from healthcheck' -a '(__fish_sn_complete)' -d 'Find and fix account data errors'
complete -c sn -n '__fish_seen_subcommand_from migrate' -a '(__fish_sn_complete)' -d 'Migrate notes to other applications'
complete -c sn -n '__fish_seen_subcommand_from organize' -a '(__fish_sn_complete)' -d 'Organize notes with AI'
complete -c sn -n '__fish_seen_subcommand_from register' -a '(__fish_sn_complete)' -d 'Register new user'
complete -c sn -n '__fish_seen_subcommand_from resync' -a '(__fish_sn_complete)' -d 'Resync content'
complete -c sn -n '__fish_seen_subcommand_from search find' -a '(__fish_sn_complete)' -d 'Search notes'
complete -c sn -n '__fish_seen_subcommand_from session' -a '(__fish_sn_complete)' -d 'Manage sessions'
complete -c sn -n '__fish_seen_subcommand_from stats' -a '(__fish_sn_complete)' -d 'Show statistics'
complete -c sn -n '__fish_seen_subcommand_from task' -a '(__fish_sn_complete)' -d 'Manage tasks'
complete -c sn -n '__fish_seen_subcommand_from tag' -a '(__fish_sn_complete)' -d 'Tag items'
complete -c sn -n '__fish_seen_subcommand_from template tpl' -a '(__fish_sn_complete)' -d 'Manage templates'
complete -c sn -n '__fish_seen_subcommand_from wipe' -a '(__fish_sn_complete)' -d 'Delete all content'
