FROM golang:1.24-alpine AS builder
ENV GOTOOLCHAIN=auto
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/ankiweb-mcp ./cmd/ankiweb-mcp

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/ankiweb-mcp /usr/local/bin/ankiweb-mcp
ENV HOME=/tmp
ENTRYPOINT ["/usr/local/bin/ankiweb-mcp"]
