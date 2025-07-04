# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod .
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} go build -o /out/app ./cmd

FROM alpine:3.22
RUN mkdir -p /app/logs
WORKDIR /app
COPY --from=builder /out/app /app/service
EXPOSE 8080
ENTRYPOINT ["/app/service"]
