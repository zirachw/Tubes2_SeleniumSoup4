# build Go binary
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /go/bin/Tubes2_SeleniumSoup4 ./cmd/server

# final image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
ARG PORT=8080
ARG CORS_ALLOWED_ORIGIN="*"
ENV PORT=${PORT}
ENV CORS_ALLOWED_ORIGIN=${CORS_ALLOWED_ORIGIN}
COPY --from=builder /go/bin/Tubes2_SeleniumSoup4 /usr/local/bin/Tubes2_SeleniumSoup4
EXPOSE 8080
ENTRYPOINT ["Tubes2_SeleniumSoup4"]