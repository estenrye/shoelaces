FROM --platform=$BUILDPLATFORM golang:1.19-alpine AS build-env
COPY internal /app/internal
COPY vendor /app/vendor
COPY go.mod go.sum main.go /app/

WORKDIR /app
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -a -ldflags='-s -w -extldflags "-static"' -o /tmp/shoelaces ./main.go

# FROM golang:1.15-alpine AS build

# WORKDIR /shoelaces
# COPY . .

# RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w -extldflags "-static"' -o /tmp/shoelaces . && \
# printf "---\nnetworkMaps:\n" > /tmp/mappings.yaml

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
