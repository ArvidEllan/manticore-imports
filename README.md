# Manticore Import Platform Starter Kit

Serverless Go starter kit for `manticore.com` using:
- AWS Lambda (`provided.al2023`)
- API Gateway HTTP API
- DynamoDB
- S3
- SES
- Serverless Framework v4

## Features included
- Health endpoint
- Public quote submission
- Public status lookup by reference + email
- Presigned upload URL generation
- Admin request listing
- Admin request details
- Admin status update with audit trail
- Static frontend pages
- Makefile and build scripts

## API endpoints
- `GET /health`
- `POST /public/quotes`
- `GET /public/status/{reference}?email=<email>`
- `POST /public/uploads/presign`
- `POST /admin/auth/login`
- `GET /admin/requests`
- `GET /admin/requests/{id}`
- `PATCH /admin/requests/{id}/status`

## Environment variables
These are supplied by `serverless.yml` in AWS. For local testing, export them manually.

- `AWS_REGION`
- `STAGE`
- `REQUESTS_TABLE`
- `AUDIT_TABLE`
- `DOCUMENTS_TABLE`
- `DOCUMENTS_BUCKET`
- `SES_FROM_EMAIL`
- `ADMIN_USERNAME`
- `ADMIN_PASSWORD`
- `JWT_SECRET`

## Local build
```bash
make build
make package
```

## Deploy
```bash
npm install -g serverless
sls deploy --stage dev
sls deploy --stage prod
```

## Notes
- Admin auth is intentionally minimal for MVP. Replace with Cognito or a real identity provider before production.
- Public status lookup requires both reference and email.
- Bucket remains private; uploads are done via presigned URLs.
- DynamoDB schema uses a single partition key per item plus GSIs for lookup.

## Suggested next steps
- Add Cognito
- Add CloudFront + frontend deployment automation
- Add WAF and rate limiting
- Add HTML email templates
- Add pagination and metrics endpoints
