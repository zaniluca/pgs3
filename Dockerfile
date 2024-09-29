FROM golang:alpine AS build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pgs3 .

FROM alpine:latest

COPY --from=build pgs3 .

ENTRYPOINT ["./pgs3"]