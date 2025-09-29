# Seif - one time secrets drop

## v0.4.0

Seif provides a simple web interface for creating secure, one-time, time-limited links to share passwords, API keys, or other sensitive data. Once accessed or expired, the secret is permanently deleted.

It's a single executable, Go is used for the server side and Svelte on the frontend.

### Goals

- **One-time access**: Secrets are automatically deleted after being viewed
- Strong encryption, **zero trust** security model
- **OAUTH2 support** for authentication - auth users can create secrets, anyone with the link can read them
- If auth is enabled, it's possible to generate **Single Use Access Links** to enable external, non-authenticated users to send a secret
- **Time-based expiration**: Configure retention periods (1-N days)
- **Size limits**: Configurable maximum secret size
- **Small code** to be able to easily review it
- **Light** on CPU, memory and bandwidth
- All the state is in the database (using `etcd-io/bbolt`)

### Non-goals

- No HTTPS (__DO USE A REVERSE PROXY__)
- No backup of the database (use cron and copy the db file)
- No rate limiting (best done via reverse proxy)
- Logged in sessions won't survive restarts

## Running

The executable is simply ran like `./seif[.exe]`. It's configured via environment variables, to be docker-friendly.

| Variable                 | Type   | Meaning                                                      | Default     |
| ------------------------ | ------ | ------------------------------------------------------------ | ----------- |
| `SEIF_DB`                | string | The path of the database                                     | `./seif.db` |
| `SEIF_PORT`              | number | Port                                                         | `34543`     |
| `SEIF_MAX_DAYS`          | number | Maximum retention days to allow                              | `3`         |
| `SEIF_DEFAULT_DAYS`      | number | Default retention days to allow, proposed in GUI             | `3`         |
| `SEIF_MAX_BYTES`         | number | Maximum size, in bytes, of a secret                          | `1024`      |
| `SEIF_OAUTH_ENABLED`     | bool   | Enable OAuth2 authentication for secret creation (see below) | `false`     |
| `SEIF_ALLOW_ACCESS_LINK` | bool   | Enable auth bypass link generation, if auth is enabled       | `false`     |

### OAuth2 Authentication (Optional)

Seif can optionally require OAuth2 authentication for creating secrets, while keeping secret retrieval publicly accessible. This is useful for organizations that want to control who can create secrets but still allow easy sharing.

When OAuth2 is enabled:
- **Secret creation** requires authentication
- **Secret retrieval** remains public (no authentication needed)
- **Secret status checking** remains public

OAuth2 works with any OpenID Connect-compatible provider (PocketID, Keycloak, Auth0, Okta, etc.).

| Variable                     | Type   | Meaning                                                           | Default                |
| ---------------------------- | ------ | ----------------------------------------------------------------- | ---------------------- |
| `SEIF_OAUTH_ENABLED`         | bool   | Enable OAuth2 authentication for secret creation                  | `false`                |
| `SEIF_OAUTH_CLIENT_ID`       | string | OAuth2 client ID from your provider                               | -                      |
| `SEIF_OAUTH_CLIENT_SECRET`   | string | OAuth2 client secret from your provider                           | -                      |
| `SEIF_OAUTH_REDIRECT_URI`    | string | OAuth2 callback URL (`http://host:port/api/auth/callback`)        | -                      |
| `SEIF_OAUTH_AUTH_URL`        | string | OAuth2 authorization endpoint                                     | -                      |
| `SEIF_OAUTH_TOKEN_URL`       | string | OAuth2 token endpoint                                             | -                      |
| `SEIF_OAUTH_USERINFO_URL`    | string | OAuth2 user info endpoint                                         | -                      |
| `SEIF_OAUTH_SCOPES`          | string | OAuth2 scopes (space-separated)                                   | `openid email profile` |
| `SEIF_OAUTH_EMAIL_WHITELIST` | string | Comma-separated list of allowed email addresses (empty=allow all) | -                      |

### OAuth2 Configuration Example

To enable OAuth2 authentication with PocketID:

```bash
export SEIF_OAUTH_ENABLED=true
export SEIF_OAUTH_CLIENT_ID=your_client_id
export SEIF_OAUTH_CLIENT_SECRET=your_client_secret
export SEIF_OAUTH_REDIRECT_URI=http://localhost:34543/api/auth/callback
export SEIF_OAUTH_AUTH_URL=https://pocketid.io/oauth/authorize
export SEIF_OAUTH_TOKEN_URL=https://pocketid.io/oauth/token
export SEIF_OAUTH_USERINFO_URL=https://pocketid.io/oauth/userinfo
export SEIF_OAUTH_EMAIL_WHITELIST="admin@company.com,user@company.com"
export SEIF_ALLOW_ACCESS_LINK=true
./seif
```

**Important**:
- Register the callback URL (`http://host:port/api/auth/callback`) in your OAuth2 provider's application settings
- Omit `SEIF_OAUTH_EMAIL_WHITELIST` to allow any authenticated user, or set it to restrict access to specific email addresses

### Access Links (Optional)

When OAuth2 is enabled, authenticated users can optionally generate access links that allow unauthenticated users to create secrets temporarily. This feature is controlled by the `SEIF_ALLOW_ACCESS_LINK` environment variable.

**How Access Links Work**:
- **Generation**: Authenticated users can create single-use, time-limited links (1-24 hours)
- **Usage**: Recipients can use these links to create secrets without authentication
- **Security**: Each link can only be used once and expires automatically if not used

**Configuration**:
```bash
export SEIF_OAUTH_ENABLED=true # prerequisite
...
export SEIF_ALLOW_ACCESS_LINK=true
```

**Use Cases**:
- Sharing secret creation access with external collaborators
- Temporary access for users without OAuth2 accounts

**Security Considerations**:
- Links are single-use and automatically expire
- Generation is logged for audit purposes
- Feature is opt-in

## Installing (with Docker)

Simple install, with docker:

`docker run --rm -i --user 1000:1000 -p 34543:34543 -e SEIF_MAX_BYTES=2048 -v seif:/data ghcr.io/proofrock/seif:latest`

Docker images for AMD64 and AARCH64 are in the 'Packages' section of this repository.

Notes when running the docker image:
- Don't change the port (env var `SEIF_PORT`), unless you know what you're doing. The healthcheck will fail. Use port remapping instead.
- The `SEIF_DB` env var value is `/data/seif.db` by default in the image; if you change it, also adjust the volume mapping (`/data` in the example above) to suit the new value.
