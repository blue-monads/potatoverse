# Cloudy

Cloudy is a multi-tenant server for Potatoverse. Main tenant user info is stored in one database, and each tenant has its own database for potatoverse app states. It is using Turso as the database.

## Auth

Branca tokens (`github.com/hako/branca`) after `POST /zz/cloudy/login`.
Send `Authorization: Bearer <token>`.

- `utype=normal` (default) — own tenant only
- `utype=admin` — list/add users, reset passwords

## Public

- `POST /zz/cloudy/sign-up`
- `GET /zz/cloudy/verify?token=...`
- `POST /zz/cloudy/login`
- `GET /zz/cloudy/health`

## Authed

- `GET /zz/cloudy/me`
- `POST /zz/cloudy/apps/:name/load`
- `GET /zz/cloudy/apps/:name/open`

## Admin

- `GET /zz/cloudy/users`
- `POST /zz/cloudy/users` — add user (`utype`, `verified`, …)
- `POST /zz/cloudy/users/:id/reset-password`

## Mail

`net/smtp` directly via `SMTP_*` envs (logs only if `SMTP_HOST` is empty).
