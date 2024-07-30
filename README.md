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

## Usage
```
golinks: a simple self-hosted implementation of go links for use in a self-
hosted environment.

Usage: golinks [-port 8080] [-config ./links]

-h                                      Show this help message
-port <number>                          The port to listen on (default: 8080)
-storage <FILE|NONE>                    The type of storage to use for persistence.
                                        Defaults to "FILE". Storage types:
                                            * NONE: Provides no persistence
                                            * FILE: Persists shortcut entries to the
                                                    file specified by the -config option
-config <absolute path to config file>  The path to the preferred config file.
                                        If this file is not present, falls back
                                        to default locations in the following
                                        order:
                                            * "./links"
                                            * "~/.config/golinks/links"
                                            * "/etc/golinks/links"
-level <loglevel>                       The loglevel to log at. Defaults to "INFO"

Config format:
The config file is a simple plaintext file consisting of one key/value pair per
line, separated by spaces, like so:

    test https://www.google.com

The value of the pair must be a full web address. Query params are not
respected, though full paths are.
```
