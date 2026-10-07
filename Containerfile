ARG BUILD_VERSION=unknown
ARG BUILD_GIT_REF=unknown
ARG BUILD_TIMESTAMP=unknown
ARG BUILD_SHORT_REF=unknown

FROM golang:1.27 AS builder
ARG BUILD_VERSION
ARG BUILD_GIT_REF
ARG BUILD_TIMESTAMP
ARG TARGETARCH
ENV BUILD_VERSION=${BUILD_VERSION}
ENV BUILD_GIT_REF=${BUILD_GIT_REF}
ENV BUILD_TIMESTAMP=${BUILD_TIMESTAMP}
WORKDIR /app
COPY . .
RUN set -eu; \
    arch="${TARGETARCH:-$(uname -m)}"; \
    case "$arch" in \
      amd64|x86_64) tailwind_arch=x64 ;; \
      arm64|aarch64) tailwind_arch=arm64 ;; \
      *) echo "unsupported architecture: $arch" >&2; exit 1 ;; \
    esac; \
    wget -nv -O /usr/bin/tailwindcss "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-${tailwind_arch}"; \
    chmod +x /usr/bin/tailwindcss
RUN set -eu; \
    arch="${TARGETARCH:-$(uname -m)}"; \
    case "$arch" in \
      amd64|x86_64) yq_arch=amd64 ;; \
      arm64|aarch64) yq_arch=arm64 ;; \
      *) echo "unsupported architecture: $arch" >&2; exit 1 ;; \
    esac; \
    wget -nv -O /usr/bin/yq "https://github.com/mikefarah/yq/releases/latest/download/yq_linux_${yq_arch}"; \
    chmod +x /usr/bin/yq
RUN wget -nv -O- https://golangci-lint.run/install.sh | sh -s -- -b /usr/bin
RUN wget -nv -O- https://taskfile.dev/install.sh | sh -s -- -b /usr/bin
RUN go mod download -x
RUN task generate fmt build-static build-healthcheck

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
    org.opencontainers.image.licenses="AGPL-3.0" \
    org.opencontainers.image.title="Account Center" \
    org.opencontainers.image.description="Self-hosted, OIDC-authenticated portal for internal services and knowledge base articles." \
    org.opencontainers.image.base.name="gcr.io/distroless/static:nonroot"
ENV AC_CATALOG_PATH=/data/catalog.yaml
ENV AC_KB_PATH=/data/kb
WORKDIR /app
COPY --from=builder /app/bin/account-center .
COPY --from=builder /app/bin/healthcheck .
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD ["/app/healthcheck"]
ENTRYPOINT ["/app/account-center"]
