# sn-cli
a command-line interface for [Standard Notes](https://standardnotes.org/).

[![Build Status](https://www.travis-ci.org/jonhadfield/sn-cli.svg?branch=master)](https://www.travis-ci.org/jonhadfield/sn-cli) [![Go Report Card](https://goreportcard.com/badge/github.com/jonhadfield/sn-cli)](https://goreportcard.com/report/github.com/jonhadfield/sn-cli)

## latest updates

### version 0.3.4 - 2024-01-07

- fix command completion and update instructions

### version 0.3.3 - 2024-01-07

- add `task` command for management of Checklists and Advanced Checklists

### version 0.3.2 - 2024-01-06

- bug fixes and sync speed increases

### version 0.3.1 - 2023-12-20

- various output improvements, including stats

### version 0.3.0 - 2023-12-14

- bug fixes and item schema tests

### version 0.2.8 - 2023-12-07

- stored sessions are now auto-renewed when expired, or nearing expiry

### version 0.2.7 - 2023-12-06

- various release packaging updates - thanks: [@clayrosenthal](https://github.com/clayrosenthal)



## current features

```
COMMANDS:
     add        add items
     delete     delete items
     edit       edit notes
     tag        tag items
     task       manage checklists and tasks
     session    store session to
     register   register an account
     resync     delete and repopulate cache
     get        get item data
     stats      show statistics
     wipe       deletes all tags and notes
     test-data  create test data (hidden option)
```
*note: export and import currently disabled due to recent StandardNotes API changes*

## installation
Download the latest release here: https://github.com/jonhadfield/sn-cli/releases

### macOS and Linux

Install:
``` console
$ install <sn-cli binary> /usr/local/bin/sn
```

### Windows

An installer is planned, but for now...
Download the binary 'sncli_windows_amd64.exe' and rename to sn.exe

## running

To see commands and options:
``` console
$ sn --help
```
### authentication

By default, your credentials will be requested every time, but you can store them using either environment variables or, on MacOS and Linux, store your session using the native Keychain application.

#### environment variables
Note: if using 2FA, the token value will be requested each time
``` shell
export SN_EMAIL=<email address>
export SN_PASSWORD=<password>
export SN_SERVER=<https://myserver.example.com>   # optional, if running personal server
```

#### session (macOS Keychain / Gnome Keyring)
Using a session is different from storing credentials as you no longer need to authenticate. As a result, if using 2FA (Two Factor Authentication), you won't need to enter your token value each time.
##### add session
```
sn session --add   # session will be stored after successful authentication
```
To encrypt your session when adding:
```
sn session --add --session-key   # either enter key as part of command, or '.' to hide its input
```
##### using a session
Prefix any command with ```--use-session``` to automatically retrieve and use the session.
If your session is encrypted, you will be prompted for the session key. To specify the key on the command line:
```
sn --use-session --session-key <key> <command>
```
To use your session automatically, set the environment variable ```SN_USE_SESSION``` to ```true```

## known issues

- accounts registered via sn-cli are initialised without initial encryption key(s). The workaround is to log in via the offical web/desktop app, to create these keys, after initial registration.

## bash autocompletion

#### tool
the bash completion tool should be installed by default on most Linux installations.

To install on macOS (Homebrew)
``` console
$ brew install bash_completion
```
then add the following to ~/.bash_profile:
``` bash
[ -f /usr/local/etc/bash_completion ] && . /usr/local/etc/bash_completion
```
#### installing completion script ([found here](https://github.com/jonhadfield/sn-cli/tree/master/autocomplete/bash_autocomplete))
##### macOS
``` console
$ cp bash_autocomplete /usr/local/etc/bash_completion.d/sn
$ echo "source /usr/local/etc/bash_completion.d/sn" | tee -a ~/.bashrc
```
##### Linux
``` console
$ cp bash_autocomplete /etc/bash_completion.d/sn
$ echo "source /etc/bash_completion.d/sn" | tee -a ~/.bashrc
```

##### autocomplete commands
``` console
$ sn <tab>
```
