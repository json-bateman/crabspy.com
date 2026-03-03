FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/a-h/templ/cmd/templ@latest
COPY . .
RUN templ generate
RUN go build -o crabspy ./cmd/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/crabspy .
RUN mkdir -p /app/data

EXPOSE 3012

CMD ["./crabspy"]
