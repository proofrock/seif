# docker buildx build --no-cache -t germanorizzo/seif:v0.0.2 . --push
# docker run --rm -i -p 34543:34543 --user 1000:1000 -v seif:/data germanorizzo/seif:v0.0.2

FROM node:latest AS build-fe

WORKDIR /app
COPY . .

RUN make build-frontend

FROM golang:latest AS build-be

WORKDIR /go/src/app
COPY . .
COPY --from=build-fe /app/backend/static ./backend/static

RUN make build-backend-nostatic

# Now copy it into our base image.
FROM debian:stable-slim

COPY --from=build-be /go/src/app/bin/seif /

RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

VOLUME /data

ENV SEIF_DB="/data/seif.db"
ENV SEIF_PORT=34543
ENV SEIF_MAX_DAYS=3
ENV SEIF_DEFAULT_DAYS=3
ENV SEIF_MAX_BYTES=1024

EXPOSE 34543

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl http://localhost:34543/api/ping || exit 1

ENTRYPOINT ["/seif"]