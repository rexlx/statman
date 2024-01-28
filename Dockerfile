FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/statman .

# Path: Dockerfile
FROM alpine:latest

RUN apk --no-cache add ca-certificates
    # addgroup -S statman && \
    # adduser -S statman -G statman

# RUN chown -R statman:statman /app


# USER statman

COPY --from=builder /app/statman /app/statman

EXPOSE 20080

CMD ["/app/statman"]
