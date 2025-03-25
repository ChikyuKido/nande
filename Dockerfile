#Build stage
FROM golang:1.24.1-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY . .

RUN cd extensions && ./build-extensions.sh && cd ..
RUN mv extension-build default-extensions

COPY go.mod go.sum ./
RUN go mod download

RUN go build -ldflags="-s -w"

#Run stage
FROM alpine:latest

RUN apk add --no-cache docker-cli hdparm smartmontools lsblk && rm -rf /var/cache/apk/*

ENV WEB_PORT=6643
ENV EXTENSION_FOLDER=extensions

COPY --from=builder /app/nande /app/nande
COPY --from=builder /app/default-extensions /app/default-extensions
COPY docker/start.sh /app/start.sh
RUN chmod +x /app/start.sh
WORKDIR /app
ENTRYPOINT ["/app/start.sh"]
