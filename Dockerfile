FROM golang:alpine AS build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pg-s3-toolkit .

FROM alpine:latest

COPY --from=build pg-s3-toolkit .

ENTRYPOINT ["./pg-s3-toolkit"]