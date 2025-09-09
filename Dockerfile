ARG base_image=cgr.dev/chainguard/wolfi-base
ARG builder_image=golang

ARG BUILDPLATFORM
FROM --platform=${BUILDPLATFORM} ${builder_image} AS builder

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

COPY . /src
WORKDIR /src
ENV CGO_ENABLED=0
RUN go mod download
RUN go build -o /assets/in ./cmd/in
RUN go build -o /assets/out ./cmd/out
RUN go build -o /assets/check ./cmd/check
RUN set -e; for pkg in $(go list ./...); do \
    go test -o "/tests/$(basename $pkg).test" -c $pkg; \
    done

FROM ${base_image} AS resource
LABEL org.opencontainers.image.url="https://ghcr.io/pixelairio/github-resource"
LABEL org.opencontainers.image.source="https://github.com/PixelAirIO/github-resource"
LABEL org.opencontainers.image.authors="Pixel Air IO https://pixelair.io"
LABEL org.opencontainers.image.vendor="Pixel Air IO"
COPY --from=builder assets/ /opt/resource/

FROM resource AS tests
COPY --from=builder /tests /tests
RUN set -e; for test in /tests/*.test; do \
    $test; \
    done
