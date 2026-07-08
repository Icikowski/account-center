ARG BUILD_VERSION=unknown
ARG BUILD_GIT_REF=unknown
ARG BUILD_TIMESTAMP=unknown
ARG BUILD_SHORT_REF=unknown

FROM golang:1.26 AS builder
ARG BUILD_VERSION
ARG BUILD_GIT_REF
ARG BUILD_TIMESTAMP
ENV BUILD_VERSION=${BUILD_VERSION}
ENV BUILD_GIT_REF=${BUILD_GIT_REF}
ENV BUILD_TIMESTAMP=${BUILD_TIMESTAMP}
WORKDIR /app
COPY . .
RUN wget -O /usr/bin/tailwindcss "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64"
RUN chmod +x /usr/bin/tailwindcss
RUN go install github.com/mikefarah/yq/v4@latest
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
RUN go install github.com/go-task/task/v3/cmd/task@latest
RUN go mod download -x
RUN task generate fmt build-static

FROM gcr.io/distroless/static:nonroot
ARG BUILD_VERSION
ARG BUILD_GIT_REF
ARG BUILD_TIMESTAMP
ARG BUILD_SHORT_REF
LABEL org.opencontainers.image.created="${BUILD_TIMESTAMP}" \
      org.opencontainers.image.authors="Piotr Icikowski" \
      org.opencontainers.image.url="https://sr.ht/~icikowski/account-center" \
      org.opencontainers.image.documentation="https://git.sr.ht/~icikowski/account-center/blob/${BUILD_SHORT_REF}/README_CONTAINER.md" \
      org.opencontainers.image.source="https://git.sr.ht/~icikowski/account-center" \
      org.opencontainers.image.version="${BUILD_VERSION}" \
      org.opencontainers.image.revision="${BUILD_GIT_REF}" \
      org.opencontainers.image.vendor="Piotr Icikowski" \
      org.opencontainers.image.licenses="GPL-3.0-only" \
      org.opencontainers.image.title="Account Center" \
      org.opencontainers.image.description="Self-hosted, OIDC-authenticated portal for internal services and knowledge base articles." \
      org.opencontainers.image.base.name="gcr.io/distroless/static:nonroot"
ENV AC_CATALOG_PATH=/data/catalog.yaml
ENV AC_KB_PATH=/data/kb
WORKDIR /app
COPY --from=builder /app/bin/account-center .
ENTRYPOINT ["/app/account-center"]
