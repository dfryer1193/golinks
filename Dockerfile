FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:latest as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR .
COPY . .

ENV CGO_ENABLED 0
ENV GOOS ${TARGETOS}
ENV GOARCH ${TARGETARCH}

RUN go build -ldflags="-s -w" -o /bin/golinks

FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch as run

WORKDIR /golinks
COPY --from=builder /bin/golinks /golinks

ENTRYPOINT ["/golinks/golinks"]
EXPOSE 8080
