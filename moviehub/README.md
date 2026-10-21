# MovieHub API

## Run locally

Set `DATABASE_URL` to a PostgreSQL connection string and `JWT_SECRET` to a
random value of at least 32 characters. Optionally set
`ALLOWED_ORIGINS` to a comma-separated list of exact frontend origins; it
defaults to no allowed browser origins. The server listens on `PORT`, or
`8080` when `PORT` is unset.

```sh
cd moviehub
DATABASE_URL='postgresql://postgres:secret@localhost:5432/moviehub?sslmode=disable' \
JWT_SECRET='replace-with-a-random-secret-at-least-32-characters' \
ALLOWED_ORIGINS='http://localhost:3000' \
go run .
```

## Deploy on Render

Create a Go Web Service connected to this repository with:

- Root Directory: `moviehub`
- Build Command: `go build -o app .`
- Start Command: `./app`
- Health Check Path: `/api/movies`

Set `DATABASE_URL` to the Supabase **Session pooler** URI (port `5432`) with
`sslmode=require`, `JWT_SECRET` to a new random value of at least 32
characters, and `ALLOWED_ORIGINS` to the exact deployed frontend origin(s),
comma-separated. Render supplies `PORT`; do not set it manually.

After loading the schema into Supabase, run [`enable_rls.sql`](./enable_rls.sql)
in its SQL Editor before connecting the database to a public frontend. Keep
database credentials and signing secrets in the hosting provider's environment
settings, not in source control.
