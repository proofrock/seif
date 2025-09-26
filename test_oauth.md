# OAuth2 Implementation Test Guide

## Testing OAuth2 Disabled (Default)

1. Run the application without OAuth2 environment variables:
   ```bash
   cd backend
   go run .
   ```
   You should see: `OAuth2: Disabled (SEIF_OAUTH_ENABLED is not set to 'true')`

2. Open http://localhost:8080 in your browser
3. Verify that:
   - The application works normally (no login required)
   - You can create secrets without authentication
   - No login/logout buttons are shown in the UI

## Testing OAuth2 Configuration Validation

1. Try enabling OAuth2 without proper configuration:
   ```bash
   SEIF_OAUTH_ENABLED=true go run .
   ```
   You should see error messages showing which configuration variables are missing.

## Testing OAuth2 Enabled

### PocketID Configuration Examples

#### Allow All Authenticated Users
```bash
export SEIF_OAUTH_ENABLED=true
export SEIF_OAUTH_CLIENT_ID=your_pocketid_client_id
export SEIF_OAUTH_CLIENT_SECRET=your_pocketid_client_secret
export SEIF_OAUTH_REDIRECT_URI=http://localhost:8080/api/auth/callback
export SEIF_OAUTH_AUTH_URL=https://pocketid.io/oauth/authorize
export SEIF_OAUTH_TOKEN_URL=https://pocketid.io/oauth/token
export SEIF_OAUTH_USERINFO_URL=https://pocketid.io/oauth/userinfo
export SEIF_OAUTH_SCOPES="openid email profile"
```

#### Restrict to Specific Email Addresses
```bash
export SEIF_OAUTH_ENABLED=true
export SEIF_OAUTH_CLIENT_ID=your_pocketid_client_id
export SEIF_OAUTH_CLIENT_SECRET=your_pocketid_client_secret
export SEIF_OAUTH_REDIRECT_URI=http://localhost:8080/api/auth/callback
export SEIF_OAUTH_AUTH_URL=https://pocketid.io/oauth/authorize
export SEIF_OAUTH_TOKEN_URL=https://pocketid.io/oauth/token
export SEIF_OAUTH_USERINFO_URL=https://pocketid.io/oauth/userinfo
export SEIF_OAUTH_EMAIL_WHITELIST="admin@company.com,user@company.com,manager@company.com"
export SEIF_OAUTH_SCOPES="openid email profile"
```

2. Run the application:
   ```bash
   cd backend
   go run .
   ```
   You should see: `OAuth2: Enabled with custom provider`

3. Open http://localhost:8080 in your browser
4. Verify that:
   - When visiting the page, an elegant authentication message appears with lock icon
   - The message says "Please login to create secrets" where "login" is a clickable link
   - Clicking "login" redirects to the OAuth provider
   - After successful authentication, user info appears in the navbar
   - You can create secrets after authentication
   - "Logout" icon button with tooltip works correctly
   - Secret retrieval (show secret) still works without authentication

## Testing Email Whitelist

1. **Configure with email whitelist** containing your test email:
   ```bash
   export SEIF_OAUTH_EMAIL_WHITELIST="your-test-email@domain.com"
   ```

2. **Test allowed email**: Login with an email in the whitelist - should work normally

3. **Test denied email**: Login with an email NOT in the whitelist - should see:
   - Server log: `OAuth2: Access denied for email: [email] (not in whitelist)`
   - Browser: User-friendly "Access Denied" page with option to return home

4. **Test no whitelist**: Remove `SEIF_OAUTH_EMAIL_WHITELIST` entirely - any authenticated email should work

## Supported Providers

Seif supports any OpenID Connect-compatible OAuth2 provider including:
- **PocketID**
- **Keycloak**
- **Auth0**
- **Okta**
- **Custom providers**

## Configuration Variables

- `SEIF_OAUTH_ENABLED`: `true` or `false` (default: `false`)
- `SEIF_OAUTH_CLIENT_ID`: OAuth client ID from your provider
- `SEIF_OAUTH_CLIENT_SECRET`: OAuth client secret from your provider
- `SEIF_OAUTH_REDIRECT_URI`: Callback URL (usually `http://localhost:8080/api/auth/callback`)
- `SEIF_OAUTH_AUTH_URL`: Authorization endpoint (e.g., `https://pocketid.io/oauth/authorize`)
- `SEIF_OAUTH_TOKEN_URL`: Token endpoint (e.g., `https://pocketid.io/oauth/token`)
- `SEIF_OAUTH_USERINFO_URL`: User info endpoint (e.g., `https://pocketid.io/oauth/userinfo`)
- `SEIF_OAUTH_SCOPES`: Space-separated scopes (optional, defaults to `openid email profile`)
- `SEIF_OAUTH_EMAIL_WHITELIST`: Comma-separated email addresses (optional, if omitted allows all)

## API Endpoints

- `GET /api/auth/login` - Initiate OAuth2 login
- `GET /api/auth/callback` - OAuth2 callback handler
- `POST /api/auth/logout` - Logout user
- `GET /api/auth/user` - Get current user info

## Protected Endpoints

- `PUT /api/putSecret` - Requires authentication if OAuth2 is enabled

## Unprotected Endpoints (as requested)

- `DELETE /api/getSecret` - Secret retrieval (show secret)
- `GET /api/getSecretStatus` - Check secret status