FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/server /bin/server
COPY --from=builder /go/bin/migrate /bin/migrate
COPY migrations /migrations
EXPOSE 8080
ENTRYPOINT ["/bin/server"]
