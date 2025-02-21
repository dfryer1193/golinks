```
 ██████╗  ██████╗     ██╗██╗     ██╗███╗   ██╗██╗  ██╗███████╗
██╔════╝ ██╔═══██╗   ██╔╝██║     ██║████╗  ██║██║ ██╔╝██╔════╝
██║  ███╗██║   ██║  ██╔╝ ██║     ██║██╔██╗ ██║█████╔╝ ███████╗
██║   ██║██║   ██║ ██╔╝  ██║     ██║██║╚██╗██║██╔═██╗ ╚════██║
╚██████╔╝╚██████╔╝██╔╝   ███████╗██║██║ ╚████║██║  ██╗███████║
 ╚═════╝  ╚═════╝ ╚═╝    ╚══════╝╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝
```
# golinks
A simple, self-hosted golinks implementation

## What are golinks?
Go links (or golinks or go/links) are browser-based redirects allowing mnemonic bookmarks or shortcuts. To use them, navigate to `http://go/<shortcut>`

For example, to get to reddit, I might set up the go/link `go/red`, which would redirect me to `https://reddit.com`.

## Client Setup
Users must visit `http://go` at least once before the browser will recognize the server as a valid address.

## Server Usage
Run the server binary on your server. In order to ensure that the server is recognized by the browser, ensure that port 80 is connected to the server in some way, either through the use of `-port 80` or by mapping port 80 to the docker container the service is running in.

Next, configure DNS to ensure that `go` points at the IP address of the hosting server.

Make sure that the address the server lives at is not publicly accessible, or anyone will be able to change your golinks.

## Help Text
```
golinks: a simple self-hosted implementation of go links for use in a self-
hosted environment.

Usage: golinks [-port 8080] [-config ./links]

-h                                      Show this help message
-port <number>                          The port to listen on (default: 8080)
-storage <FILE|NONE>                    The type of storage to use for
                                        persistence. Defaults to "FILE". Storage
                                        types:
                                            * NONE: Provides no persistence
                                            * FILE: Persists shortcut entries to
                                                    the file specified by the
                                                    -config option
-config <absolute path to config file>  The path to the preferred config file.
                                        If this file is not present, falls back
                                        to default locations in the following
                                        order:
                                            * "./links"
                                            * "~/.config/golinks/links"
                                            * "/etc/golinks/links"
-level <loglevel>                       The loglevel to log at. Defaults to
                                        "INFO"

Config format:
The config file is a simple plaintext file consisting of one key/value pair per
line, separated by spaces, like so:

    test https://www.google.com

The value of the pair must be a full web address. Query params are not
respected, though full paths are.
```
