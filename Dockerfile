#Build stage
FROM golang:1.24.1-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY . .

COPY go.mod go.sum ./
RUN go mod download

RUN go build -ldflags="-s -w"

#Run stage
FROM golang:1.24.1-alpine

RUN apk add --no-cache docker hdparm smartmontools lsblk

ENV WEB_PORT=6643
ENV EXTENSION_FOLDER=extension-build

COPY --from=builder /app/nande /app/nande
COPY --from=builder /app/extensions /app/default-extensions
COPY docker/start.sh /app/start.sh
RUN chmod +x /app/start.sh
WORKDIR /app
ENTRYPOINT ["/app/start.sh"]
