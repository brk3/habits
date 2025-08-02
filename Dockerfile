# syntax=docker/dockerfile:1

FROM golang:1.23 AS builder

ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o habits

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/habits /habits
ENTRYPOINT ["/habits"]
