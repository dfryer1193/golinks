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
`./golinks` - runs using the included config file, listening on port 8080.

```
Usage: golinks [-port 8080] [-config ./links]

-h                                      Show this help message
-port <number>                          The port to listen on (default: 8080)
-config <absolute path to config file>  The path to the preferred config file.
                                        If this file is not present, falls back
                                        to default locations in the following
                                        order:
                                            * "./links"
                                            * "~/.config/golinks/links"
                                            * "/etc/golinks/links"

Config format:
The config file is a simple plaintext file consisting of one key/value pair per
line, separated by spaces, like so:

    test https://www.google.com

The value of the pair must be a full web address. Query params are not
respected, though full paths are.
```
