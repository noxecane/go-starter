FROM golang:1.14-alpine as builder

# Ensure ca-certficates are up to date
RUN update-ca-certificates

WORKDIR /app

COPY go.mod go.sum ./

ENV GOOS=linux

RUN go mod download & go mod verify

COPY . .

# Build the binary
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/server ./cmd/godview-starter/main.go

FROM gcr.io/distroless/base

WORKDIR /app

# Copy our static executable
COPY --from=builder /app/server .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/sql ./sql

# Run the hello binary.
ENTRYPOINT ["/app/server"]