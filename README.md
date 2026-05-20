# Manticore Import Platform Starter Kit

Serverless Go platform for `manticore.com` using:
- AWS Lambda (`provided.al2023`)
- API Gateway HTTP API with WAF and throttling
- Amazon Cognito (admin authentication)
- DynamoDB, S3, SES
- CloudFront + S3 (static frontend)
- Serverless Framework v4

## Features
- Health endpoint
- Public quote submission with HTML email notifications
- Public status lookup by reference + email
- International deal scanner across overseas marketplaces
- Presigned upload URL generation
- Admin request listing with cursor pagination
- Admin metrics dashboard data
- Admin status updates with audit trail and HTML status emails
- Cognito-backed admin auth (legacy JWT fallback for local dev)
- CloudFront frontend delivery with deploy automation
- WAF rate limiting and managed rule protection

## API endpoints

### Public
- `GET /health`
- `POST /public/quotes`
- `GET /public/status/{reference}?email=<email>`
- `POST /public/uploads/presign`
- `POST /public/deals/scan`

### Admin (Bearer token required)
- `POST /admin/auth/login`
- `GET /admin/requests?limit=25&cursor=<cursor>`
- `GET /admin/requests/{id}`
- `PATCH /admin/requests/{id}/status`
- `GET /admin/metrics`

## Authentication

Production admin auth uses **Amazon Cognito**. After deploy:

1. Create an admin user in the Cognito user pool (output `CognitoUserPoolId`).
2. Add the user to the `admins` group.
3. Login via `POST /admin/auth/login` with email/password.
4. Use the returned Cognito ID token as `Authorization: Bearer <token>`.

When `COGNITO_USER_POOL_ID` and `COGNITO_CLIENT_ID` are unset (local dev), the API falls back to legacy username/password JWT auth.

## Environment variables

Set automatically by `serverless.yml` in AWS:

| Variable | Purpose |
|----------|---------|
| `AWS_REGION` | AWS region |
| `STAGE` | Deployment stage |
| `REQUESTS_TABLE` | Import quote requests |
| `AUDIT_TABLE` | Status change audit trail |
| `DOCUMENTS_TABLE` | Document metadata |
| `DOCUMENTS_BUCKET` | S3 document uploads |
| `FRONTEND_BUCKET` | S3 static frontend bucket |
| `SES_FROM_EMAIL` | Outbound email sender |
| `COGNITO_USER_POOL_ID` | Cognito admin user pool |
| `COGNITO_CLIENT_ID` | Cognito app client |
| `COGNITO_REGION` | Cognito region |
| `CLOUDFRONT_DOMAIN` | Frontend CloudFront domain |
| `ADMIN_USERNAME` / `ADMIN_PASSWORD` | Legacy local auth fallback |
| `JWT_SECRET` | Legacy JWT signing secret |

## Local build and test

```bash
make build
make package
make test
```

## Deploy

```bash
npm install -g serverless
make deploy-dev          # API + infrastructure
make deploy-frontend-dev # Sync web/public to S3 + CloudFront invalidation
make deploy-all-dev      # Both
```

On Windows:

```powershell
.\scripts\deploy-frontend.ps1 -Stage dev
```

## Infrastructure highlights

- **Cognito**: Admin-only user pool with strong password policy and `admins` group
- **CloudFront**: S3 origin with OAC, SPA-style 404 routing, HTTPS redirect
- **WAF**: 2000 req/5min IP rate limit + AWS managed common/bad-input rules on the HTTP API
- **API throttling**: 100 req/s steady, 200 burst via API Gateway
- **HTML emails**: Branded templates for quote received and status updates

## Pagination

`GET /admin/requests` returns:

```json
{
  "items": [...],
  "nextCursor": "REQUEST#...",
  "hasMore": true
}
```

Pass `cursor` from `nextCursor` on the next request. Default `limit` is 25 (max 100).

## Metrics

`GET /admin/metrics` returns request totals and counts grouped by status.

## Notes

- Public status lookup requires both reference and email.
- Document bucket remains private; uploads use presigned URLs.
- Frontend deploy script injects `js/config.js` with the live API URL.
- Verify SES domain identity before sending HTML emails in production.
