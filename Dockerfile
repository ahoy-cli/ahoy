# Dockerfile for security testing.

FROM scratch
ENTRYPOINT ["/ahoy"]
COPY ahoy /
