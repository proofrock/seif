# Seif - one time secrets storage

## v0.3.0

```text
Usage of ./seif:
  -db string
        The path of the sqlite database (default "./seif.db")
  -default-days int
        Default retention days to allow, proposed in GUI (default 3)
  -max-bytes int
        Maximum size, in bytes, of a secret (default 1024)
  -max-days int
        Maximum retention days to allow (default 3)
  -port int
        Port (default 34543)
```

Simple install, with docker:

`docker run --rm -i -p 12321:12321 -v seif:/data ghcr.io/proofrock/seif:latest`

Docker images for AMD64 and AARCH64 are in the 'Packages' section of this repository.
