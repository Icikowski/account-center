FROM golang:1.26 AS builder
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
ENV AC_CATALOG_PATH=/data/catalog.yaml
ENV AC_KB_PATH=/data/kb
WORKDIR /app
COPY --from=builder /app/bin/account-center .
ENTRYPOINT ["/app/account-center"]
