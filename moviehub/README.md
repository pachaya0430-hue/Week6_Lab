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

## Week 12: admin API

The admin API requires a valid JWT with the `admin` role. Promote the first
admin manually in Supabase SQL Editor, replacing the email with the account
that should administer the site:

```sql
UPDATE public.members SET role = 'admin' WHERE lower(email) = lower('you@example.com');
```

Log out and sign in again after changing a role so the new role is included in
the JWT. Run [`week12_migration.sql`](./week12_migration.sql) in Supabase SQL
Editor before using admin movie forms; it widens `movies.poster_path` for
external poster URLs.

Admin API routes:

- `GET /api/me` — current signed-in member.
- `GET /api/genres` — public genre list.
- `GET /api/admin/stats` — dashboard counts and summaries.
- `GET /api/admin/members` and `PUT /api/admin/members/{id}/role` — member
  management. An admin cannot change their own role.
- `GET`, `POST`, `PUT`, and `DELETE /api/admin/movies/{id}` (POST uses
  `/api/admin/movies`) — admin movie management.

The current workspace contains the Go API only; the React front-end source
described in the course worksheet is not part of this repository.
