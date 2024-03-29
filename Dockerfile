FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/statman .
RUN chmod +x /app/statman

# Path: Dockerfile
FROM alpine:latest

RUN apk --no-cache add ca-certificates
# RUN addgroup -S statman
# RUN adduser -S -G statman statman

COPY --from=builder /app/statman /app/statman
RUN mkdir /logs
# RUN chown -R statman:statman /logs

# USER statman

EXPOSE 20080

CMD ["/app/statman"]
