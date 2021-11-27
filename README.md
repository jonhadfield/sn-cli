# sn-cli
a command-line interface for [Standard Notes](https://standardnotes.org/).

[![Build Status](https://www.travis-ci.org/jonhadfield/sn-cli.svg?branch=master)](https://www.travis-ci.org/jonhadfield/sn-cli) [![Go Report Card](https://goreportcard.com/badge/github.com/jonhadfield/sn-cli)](https://goreportcard.com/report/github.com/jonhadfield/sn-cli)

## Important 
This is an early release to add support for SN accounts using encryption version 004 so please ensure you have a backup in case of any issues caused by this app.
This is only compatible with accounts either created since November 2020, or older accounts that have backed up, then restored to force an upgrade to 004.

## current features

```
COMMANDS:
     add        add items
     delete     delete items
     tag        tag items
     get        get items
     export     export data
     import     import data
     stats      show statistics
     wipe       deletes all tags and notes
     session    manage session credentials
     test-data  create test data (hidden option)
```

## installation
Download the latest release here: https://github.com/jonhadfield/sn-cli/releases

### macOS and Linux

Install:  
``
$ install <sn-cli binary> /usr/local/bin/sn
``  

### Windows
  
An installer is planned, but for now...  
Download the binary 'sncli_windows_amd64.exe' and rename to sn.exe

## running

To see commands and options:  
``
$ sn --help
``
### authentication

By default, your credentials will be requested every time, but you can store them using either environment variables or, on MacOS and Linux, store your session using the native Keychain application.

#### environment variables
Note: if using 2FA, the token value will be requested each time
```
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

- sessions added to keychain do not currently refresh themselves. The workaround is to run re-add the session if an invalid session message is returned.
- accounts registered via sn-cli are initialised without initial encryption key(s). The workaround is to log in via the offical web/desktop app, to create these keys, after initial registration.

## bash autocompletion

#### tool
the bash completion tool should be installed by default on most Linux installations.  

To install on macOS (Homebrew)  
``
$ brew install bash_completion  
``  
then add the following to ~/.bash_profile:  
``  
[ -f /usr/local/etc/bash_completion ] && . /usr/local/etc/bash_completion
`` 
#### installing completion script ([found here](https://github.com/jonhadfield/sn-cli/tree/master/autocomplete/bash_autocomplete))
##### macOS  
``  
$ cp bash_autocomplete /usr/local/etc/bash_completion.d/sn
``  
##### Linux  
``
$ cp bash_autocomplete /etc/bash_completion.d/sn
``

##### autocomplete commands
``
$ sn <tab>
``
