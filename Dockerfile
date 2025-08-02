FROM gcr.io/distroless/static:nonroot

COPY habits /habits
ENTRYPOINT ["/habits"]
