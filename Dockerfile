FROM gcr.io/distroless/static:nonroot

ARG BIN=dist/habits-linux-arm64
COPY ${BIN} /habits
ENTRYPOINT ["/habits"]