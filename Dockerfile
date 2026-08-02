FROM gcr.io/distroless/static:nonroot

ARG TARGETARCH
COPY dist/habits-linux-${TARGETARCH} /habits
ENTRYPOINT ["/habits"]
