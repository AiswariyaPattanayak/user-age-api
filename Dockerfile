FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -o /bin/server ./cmd/server

FROM alpine:3.20
RUN adduser -D app
USER app
COPY --from=builder /bin/server /bin/server
EXPOSE 3000
ENTRYPOINT ["/bin/server"]
