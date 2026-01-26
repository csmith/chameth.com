# Chameth.com

This project contains the code for the personal website of Chris Smith.
The project is written in Go, backed by a postgresql database, and uses
go templates for rendering content.

## Project structure

- `cmd/serve` - main program code
- `cmd/serve/db` - database handling code
- `cmd/serve/admin` - admin interface. this is is exposed on tailscale,
  accessible only by the site admin

## Common patterns

### Paths

Where content is accessible via a path, its database table should
have a `path` column, and then triggers should be created to
automatically populate/update the `paths` table when the content
table is changed.

## Code standards

### General

- Use the `log/slog` package for logging when required
- This is a personal website, it doesn't need unit tests or
  public API documentation.

### Database

- ID fields in database tables should always be
  `INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY`.
- Return type structs should be defined in the `models.go`
  file not inline in other database files.
- Do not create "down" migrations, we only roll forwards.

## Commands

To build the app and test it compiles:

```
go build -o /tmp/serve ./cmd/serve
```

To query the database:

```
docker compose exec -T database psql -U postgres -c "SELECT * FROM films;"
```
