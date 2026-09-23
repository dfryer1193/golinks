FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o golinks cmd/golinks/golinks.go

FROM alpine:3.18
ENV PORT=8080
EXPOSE $PORT

RUN mkdir -p /config

COPY --from=builder /app/golinks /golinks

# Default to file storage unless DATABASE_URL is provided via env
CMD sh -c 'if [ -n "$DATABASE_URL" ]; then \
    /golinks -storage POSTGRES -config "$DATABASE_URL" -migrate-from /config/links || true; \
    exec /golinks -storage POSTGRES -config "$DATABASE_URL"; \
else \
    exec /golinks -storage FILE -config /config/links; \
fi'
