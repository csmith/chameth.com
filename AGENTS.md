# Chameth.com

This project contains the code for the personal website of Chris Smith.
The project is written in Go, backed by a postgresql database, and uses
go templates for rendering content.

## Project structure

- `cmd/serve` - main program code
- `admin` - admin interface, exposed on tailscale
- `db` - database handling code
- `external` - packages for interacting with external systems/APIs
- `templates` - frontend templates and Go template helper code

## Common patterns

### Paths

Where content is accessible via a path, its database table should
have a `path` column, and then triggers should be created to
automatically populate/update the `paths` table when the content
table is changed.

### Shortcodes

Reusable chunks of content are exposed as shortcodes, defined under
`content/shortcodes`. These are usable in dynamic content via
`{%shortcodename arg1 arg%}` or `{%shortcodename arg1%}arg2{%endshortcodename%}`
markup. They can also expose a `Render()` function when the
content is needed programatically.

### Configuration and secrets

External URLs, usernames, passwords, etc should be defined as flags.
Flags should be defined close to where they're used, but hoisted to
keep packages reusable where it makes sense. e.g. the
`external/atproto` package takes configuration, and the flags are
defined at the call site in `content`.

### Admin interface

The admin interface is only accessible by a single user. We do
not need to be concerned about parallel updates, authentication,
etc.

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
