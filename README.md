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
--storage <FILE|NONE|SQLITE|POSTGRES>   The type of storage to use for
                                        persistence. Defaults to "FILE". Storage
                                        types:
                                            * NONE: Provides no persistence
                                            * FILE: Persists shortcut entries to
                                                    the file specified by the
                                                    -config option
                                            * SQLITE: Persists shortcut entries
                                                      to a sqlite db. The path
                                                      to the db is specified by
                                                      the -config option.
                                            * POSTGRES: Persists shortcut entries
                                                        to a postgres db. The
                                                        connection string is
                                                        specified by the -config
                                                        option (e.g.
                                                        postgres://user:pass@host/db).
-config <path or env>       The path to the preferred config file.
                                        If this file is not present, falls back
                                        to default locations in the following
                                        order:
                                            * "./links"
                                            * "~/.config/golinks/links"
                                            * "/etc/golinks/links"
                                        For POSTGRES, this can be omitted if
                                        DATABASE_URL or GOLINKS_DB_URL env var
                                        is set. Defaults to "FILE".
-level <loglevel>                       The loglevel to log at. Defaults to
                                        "INFO"
-migrate-from <path to file>            Migrate links from a file-based config
                                        to the configured database storage. Only
                                        valid when storage is SQLITE or POSTGRES.

Config format:
The config file is a simple plaintext file consisting of one key/value pair per
line, separated by spaces, like so:

    test https://www.google.com

The value of the pair must be a full web address. Query params are not
respected, though full paths are.


## Docker & Deployment
Multi-architecture (`linux/amd64`, `linux/arm64`) Docker images are pushed to the Harbor registry under the `library` project:

- **Registry Image**: `registry.werewolves.fyi/library/golinks:<tag>`
- **Build & Push**: Run `make` to build multi-arch images, create the manifest list, and push to Harbor.
- **Postgres migration**: When `DATABASE_URL` is set, the container runs a one-time migration from `/config/links` to Postgres, then starts normally. This is safe to rerun.

