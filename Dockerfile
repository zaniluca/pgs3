FROM golang:alpine AS build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pgs3 .

FROM alpine:latest

LABEL org.opencontainers.image.source=https://github.com/zaniluca/pgs3
LABEL org.opencontainers.image.description="PGS3, a toolkit for managing PostgreSQL backups and restores with S3"
LABEL org.opencontainers.image.licenses=MIT

COPY --from=build pgs3 .

ENTRYPOINT ["./pgs3"]