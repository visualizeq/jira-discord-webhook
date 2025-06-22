# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.21 AS builder
WORKDIR /src
COPY go.mod .
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} go build -o /out/app

FROM --platform=$TARGETPLATFORM alpine:3.18
WORKDIR /
COPY --from=builder /out/app /app
EXPOSE 8080
ENTRYPOINT ["/app"]

