FROM golang:1.22.0 AS build_ebook-uploader
ENV CGO_ENABLED 0
ARG BUILD_REF

# Create the service directory and the copy the module files first and then
# download the dependencies. If this doesn't change, we won't need to do this
# again in future builds.
# RUN mkdir /service
# COPY go.* /service/
# WORKDIR /service
# RUN go mod download


WORKDIR /app
COPY . .

# Build the service binary.
RUN go build -o bin/ebook-uploader -ldflags "-X main.build=${BUILD_REF}" cmd/server/main.go

# Run the Go Binary in Alpine.
FROM alpine:3.19

# api web service
EXPOSE 8000

# debug web service
EXPOSE 4000

ARG BUILD_DATE
ARG BUILD_REF

RUN addgroup -g 1000 -S gopher && \
    adduser -u 1000 -h /opt/service -G gopher -S gopher
COPY --from=build_ebook-uploader --chown=gopher:gopher /app/bin/ebook-uploader /opt/service/ebook-uploader

WORKDIR /opt/service
USER gopher
CMD ["/opt/service/ebook-uploader"]

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.title="ebook-uploader" \
      org.opencontainers.image.authors="Dave Derderian <dave@drt.sh>" \
      org.opencontainers.image.source="https://github.com/funayman/simple-ebook-uploader/tree/master" \
      org.opencontainers.image.revision="${BUILD_REF}" \
      org.opencontainers.image.vendor="Simple eBook Uploader"
