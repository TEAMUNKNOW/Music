FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY . .

# Generate/repair go.sum during the Railway build. This keeps the repo
# deployable even when go.sum was not committed.
RUN go mod tidy && CGO_ENABLED=0 go build -o /music-backend .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /music-backend /music-backend
EXPOSE 3000
CMD ["/music-backend"]
