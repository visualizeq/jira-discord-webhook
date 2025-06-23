# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.24 AS builder
WORKDIR /src
COPY go.mod .
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} go build -o /out/app ./cmd

FROM alpine:3.22
WORKDIR /
COPY --from=builder /out/app /app
EXPOSE 8080
ENTRYPOINT ["/app"]
