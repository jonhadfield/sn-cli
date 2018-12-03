# sncli
a command-line interface for [Standard Notes](https://standardnotes.org/).

[![Build Status](https://www.travis-ci.org/jonhadfield/sncli.svg?branch=master)](https://www.travis-ci.org/jonhadfield/sncli) [![Go Report Card](https://goreportcard.com/badge/github.com/jonhadfield/gosn)](https://goreportcard.com/report/github.com/jonhadfield/gosn)



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


## roadmap

features in progess:
- bash and zsh completion
- export: plaintext or encrypted
- local caching of encrypted items
- option to securely persist session between commands
- test and document for Windows users
- manage preferences

## install and run

Download the latest release here: https://github.com/jonhadfield/sncli/releases and install:

``
$ install <sncli binary> /usr/local/bin/sn
``

Then to see commands and options:  
```    
$ sn --help
```
## authentication

sncli will automatically prompt for credentials (including 2FA, if set) each time you run a command.  
Instead, you can set your email and/or password using environment variables:

Setting email and password:
```
$ export SN_EMAIL=<email_address>
$ export SN_PASSWORD=<password>
```

## using your own server

To override the Standard Notes server:
```
$ export SN_SERVER=https://<your_server_url>
```
