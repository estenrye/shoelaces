FROM --platform=$BUILDPLATFORM golang:1.19-alpine AS build-env
COPY internal /app/internal
COPY vendor /app/vendor
COPY go.mod go.sum main.go /app/

WORKDIR /app
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GODEBUG=netdns=go \
    go build -a -ldflags='-s -w -extldflags "-static"' -o /tmp/shoelaces ./main.go

# Final container has basically nothing in it but the executable
FROM scratch
COPY --from=build-env /tmp/shoelaces /shoelaces
COPY web /web
COPY configs/data-dir /data

ENV BIND_ADDR="0.0.0.0:8081" \
  BASE_URL="localhost:8081" \
  DATA_DIR="/data" \
  STATIC_DIR="/web" \
  TEMPLATE_EXTENSION=".slc" \
  MAPPINGS_FILE="mappings.yaml" \
  DEBUG="false"
EXPOSE 8081

ENTRYPOINT ["/shoelaces"]
CMD []
