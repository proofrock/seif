# docker buildx build --no-cache -t germanorizzo/seif:v0.0.2 . --push
# docker run --rm -i -p 12321:12321 -v seif:/data germanorizzo/seif:v0.0.2

FROM node:latest as build-fe

WORKDIR /app
COPY . .

RUN make build-frontend

from golang:latest as build-be

WORKDIR /go/src/app
COPY . .
COPY --from=build-fe /app/backend/static ./backend/static

RUN make build-backend-nostatic

# Now copy it into our base image.
FROM debian:stable-slim

COPY --from=build-be /go/src/app/bin/seif /

VOLUME /data

EXPOSE 34543

ENTRYPOINT ["/seif", "--db", "/data/seif.db"]