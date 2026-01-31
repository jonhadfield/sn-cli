# Fish shell completion for sncli
# Save this file to ~/.config/fish/completions/sncli.fish

function __fish_sncli_no_subcommand
    set -l cmd (commandline -opc)
    if [ (count $cmd) -eq 1 ]
        return 0
    end
    return 1
end

# Get completions from the CLI
function __fish_sncli_complete
    set -l cmd (commandline -opc)
    set -l cmd_str (string join ' ' $cmd)

    if test (count $cmd) -gt 1
        # Get subcommand completions
        eval command sncli $cmd[2..-1] --generate-bash-completion 2>/dev/null
    else
        # Get top-level command completions
        sncli --generate-bash-completion 2>/dev/null
    end
end

# Main completions
complete -c sncli -f -n __fish_sncli_no_subcommand -a '(__fish_sncli_complete)' -d 'Standard Notes CLI'

# Global options
complete -c sncli -l cachedb-dir -d 'Cache database directory'
complete -c sncli -l debug -d 'Enable debug mode'
complete -c sncli -l server -d 'Standard Notes server URL'
complete -c sncli -l session-key -d 'Session encryption key'
complete -c sncli -l use-session -d 'Use stored session'
complete -c sncli -s h -l help -d 'Show help'
complete -c sncli -s v -l version -d 'Show version'

# Command-specific completions
complete -c sncli -n '__fish_seen_subcommand_from add' -a '(__fish_sncli_complete)' -d 'Add items'
complete -c sncli -n '__fish_seen_subcommand_from backup bak' -a '(__fish_sncli_complete)' -d 'Backup operations'
complete -c sncli -n '__fish_seen_subcommand_from delete' -a '(__fish_sncli_complete)' -d 'Delete items'
complete -c sncli -n '__fish_seen_subcommand_from edit' -a '(__fish_sncli_complete)' -d 'Edit items'
complete -c sncli -n '__fish_seen_subcommand_from export exp' -a '(__fish_sncli_complete)' -d 'Export notes'
complete -c sncli -n '__fish_seen_subcommand_from get' -a '(__fish_sncli_complete)' -d 'Get items'
complete -c sncli -n '__fish_seen_subcommand_from organize' -a '(__fish_sncli_complete)' -d 'Organize notes with AI'
complete -c sncli -n '__fish_seen_subcommand_from register' -a '(__fish_sncli_complete)' -d 'Register new user'
complete -c sncli -n '__fish_seen_subcommand_from resync' -a '(__fish_sncli_complete)' -d 'Resync content'
complete -c sncli -n '__fish_seen_subcommand_from search find' -a '(__fish_sncli_complete)' -d 'Search notes'
complete -c sncli -n '__fish_seen_subcommand_from session' -a '(__fish_sncli_complete)' -d 'Manage sessions'
complete -c sncli -n '__fish_seen_subcommand_from stats' -a '(__fish_sncli_complete)' -d 'Show statistics'
complete -c sncli -n '__fish_seen_subcommand_from task' -a '(__fish_sncli_complete)' -d 'Manage tasks'
complete -c sncli -n '__fish_seen_subcommand_from tag' -a '(__fish_sncli_complete)' -d 'Tag items'
complete -c sncli -n '__fish_seen_subcommand_from template tpl' -a '(__fish_sncli_complete)' -d 'Manage templates'
complete -c sncli -n '__fish_seen_subcommand_from wipe' -a '(__fish_sncli_complete)' -d 'Delete all content'
