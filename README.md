# sn-cli
a command-line interface for [Standard Notes](https://standardnotes.org/).

[![Build Status](https://www.travis-ci.org/jonhadfield/sn-cli.svg?branch=master)](https://www.travis-ci.org/jonhadfield/sn-cli) [![Go Report Card](https://goreportcard.com/badge/github.com/jonhadfield/sn-cli)](https://goreportcard.com/report/github.com/jonhadfield/sn-cli)



## current features

```
COMMANDS:
     add        add items
     delete     delete items
     tag        tag items
     get        get items
     register   register a new user
     stats      show statistics
     wipe       deletes all tags and notes
     fixup      find and fix item issues
     test-data  create test data
```

*NOTE: This is a very early release so please take a backup using one of the official apps before using this to make any changes.
Please raise an issue if you find any problems.*

## changelog

0.0.1  initial  
0.0.2  added bash completion

## roadmap

features in progess:
- ~~bash and~~ zsh completion
- export: plaintext or encrypted
- local caching of encrypted items
- option to securely persist session between commands
- test and document for Windows users
- manage preferences

## install and run

Download the latest release here: https://github.com/jonhadfield/sn-cli/releases  
Install:  
``
$ install <sn-cli binary> /usr/local/bin/sn
``  
To see commands and options:  
``
$ sn --help
``

## authentication

sn-cli will automatically prompt for credentials (including 2FA, if set) each time you run a command.  
Instead, you can set your email and/or password using environment variables:

Setting email and password:  
``
$ export SN_EMAIL=<email_address>  
``  
``
$ export SN_PASSWORD=<password>  
``

## using your own server

To override the Standard Notes server:  
``
$ export SN_SERVER=https://<your_server_url>
``

## bash autocompletion

### autocompletion tool
bash completion depends on a tool that should be installed by default on most Linux installations.  

To install on macOS (Homebrew)  
``
$ brew install bash_completion  
``  
then add the following to ~/.bash_profile:  
``  
[ -f /usr/local/etc/bash_completion ] && . /usr/local/etc/bash_completion
`` 
### installing auto completion script ([found here](https://github.com/jonhadfield/sn-cli/tree/master/autocomplete))
#### macOS  
``  
$ cp autocomplete/bash_autocomplete /usr/local/etc/bash_completion.d/sn
``  
#### Linux  
``
$ cp autocomplete/bash_autocomplete /etc/bash_completion.d/sn
``

