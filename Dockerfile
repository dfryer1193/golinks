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

CMD ["/golinks", "-storage",  "FILE", "-config", "/config/links"]
