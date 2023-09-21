# docker buildx build --no-cache -t germanorizzo/gersuite:seif . --push
# docker run --rm -i -p 12321:12321 -v seif:/data germanorizzo/gersuite:seif

FROM node:latest as build

WORKDIR /app
COPY . .

RUN rm -rf public
RUN npm install
RUN npm run build

# Now copy it into our base image.
FROM germanorizzo/sqliterg:latest

COPY --from=build /app/dist /public
COPY --from=build /app/seif.yaml /seif.yaml

ENTRYPOINT ["/sqliterg", "--db", "/data/seif.db::/seif.yaml", "--serve-dir", "/public", "--index-file", "index.html"]