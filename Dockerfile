FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o mock-lsp-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/mock-lsp-server /usr/local/bin/
CMD ["mock-lsp-server"]
