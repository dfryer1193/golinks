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

## Storage Options

Golinks supports multiple storage backends:

### File Storage (Default)
Simple plaintext file with one key/value pair per line:
```bash
golinks -storage FILE -config ./links
```

Config format:
```
test https://www.google.com
gh https://github.com
```

### SQLite Storage
Persistent database storage with automatic connection pooling:
```bash
# Initialize schema
golinks migrate -storage SQLITE -config ./golinks.db

# Run server
golinks -storage SQLITE -config ./golinks.db
```

### PostgreSQL Storage
Production-ready database with connection pooling (25 connections):
```bash
# Using connection string
golinks migrate -storage POSTGRES -config "postgres://user:pass@localhost/golinks?sslmode=disable"
golinks -storage POSTGRES -config "postgres://user:pass@localhost/golinks?sslmode=disable"

# Using environment variable
export DATABASE_URL="postgres://user:pass@localhost/golinks?sslmode=disable"
golinks migrate -storage POSTGRES
golinks -storage POSTGRES
```

## Database Migrations

Before using SQLite or PostgreSQL storage, you must run migrations to initialize the schema:

```bash
# Run migrations
golinks migrate -storage SQLITE -config ./golinks.db
golinks migrate -storage POSTGRES -config "postgres://..."

# Check migration status
golinks migrate status -storage SQLITE -config ./golinks.db
```

Migrations are:
- **Idempotent**: Safe to run multiple times
- **Transactional**: Either fully applied or fully rolled back
- **Embedded**: No external files needed
- **Version-tracked**: Stored in `schema_migrations` table

## Environment Variables

- `DATABASE_URL`: PostgreSQL connection string (alternative to `-config` flag)

## Help Text
```
golinks: a simple self-hosted implementation of go links for use in a self-
hosted environment.

Usage: golinks [-port 8080] [-config ./links]

-h                                      Show this help message
-port <number>                          The port to listen on (default: 8080)
-storage <FILE|NONE|SQLITE|POSTGRES>    The type of storage to use for
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
                                                        to a PostgreSQL database.
                                                        Connection string via
                                                        -config option or
                                                        DATABASE_URL env var.
-config <absolute path to config file>  The path to the preferred config file.
                                        If this file is not present, falls back
                                        to default locations in the following
                                        order:
                                            * "./links"
                                            * "~/.config/golinks/links"
                                            * "/etc/golinks/links"
                                        For POSTGRES storage, this should be a
                                        connection string (or use DATABASE_URL
                                        environment variable).
-level <loglevel>                       The loglevel to log at. Defaults to
                                        "INFO"
```

## Deployment Examples

### Docker with SQLite
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o golinks ./cmd/golinks

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/
COPY --from=builder /app/golinks .

# Create data directory and initialize database
RUN mkdir -p /data && \
    ./golinks migrate -storage SQLITE -config /data/golinks.db

# Run server
VOLUME ["/data"]
EXPOSE 80
CMD ["./golinks", "-port", "80", "-storage", "SQLITE", "-config", "/data/golinks.db"]
```

### Docker Compose with PostgreSQL
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: golinks
      POSTGRES_USER: golinks
      POSTGRES_PASSWORD: changeme
    volumes:
      - postgres_data:/var/lib/postgresql/data

  golinks:
    build: .
    ports:
      - "80:80"
    environment:
      DATABASE_URL: postgres://golinks:changeme@postgres:5432/golinks?sslmode=disable
    depends_on:
      - postgres
    command: sh -c "./golinks migrate -storage POSTGRES && ./golinks -port 80 -storage POSTGRES"

volumes:
  postgres_data:
```

### Kubernetes with PostgreSQL
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: golinks
spec:
  replicas: 2
  selector:
    matchLabels:
      app: golinks
  template:
    metadata:
      labels:
        app: golinks
    spec:
      initContainers:
      - name: migrate
        image: golinks:latest
        command: ["./golinks", "migrate", "-storage", "POSTGRES"]
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: golinks-db
              key: connection-string
      containers:
      - name: golinks
        image: golinks:latest
        ports:
        - containerPort: 80
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: golinks-db
              key: connection-string
        args: ["-port", "80", "-storage", "POSTGRES"]
```
