# Chameth.com

This project contains the code for the personal website of Chris Smith.
The project is written in Go, backed by a postgresql database, and uses
go templates for rendering content.

## Project structure

- `cmd/serve` - main program code
- `cmd/generate` - code generator for shortcode/asset registrations
- `admin` - admin interface, exposed on tailscale
- `db` - shared database handling code
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
`features/shortcodes`. These are usable in dynamic content via
`{%shortcodename arg1 arg%}` or `{%shortcodename arg1%}arg2{%endshortcodename%}`
markup. They can also expose a `Render()` function when the
content is needed programatically.

Shortcodes can also be called from Go templates via the `Component`
function on `PageData`: `{{call .Component "shortcodename" arg1 arg2}}`.
This avoids pre-rendering shortcode HTML in handlers. Arguments are
converted to strings via `fmt.Sprint` so template values like ints
can be passed directly.

Any CSS file in a shortcode package will be included in the compiled
stylesheet served by the site.

### CSS

Global frontend styles are in `assets/stylesheet`. CSS must NEVER be
inlined in HTML in the frontend. Selectors should be nested where
possible/appropriate. Public CSS and JS files that should be bundled
must use the `.public.css` and `.public.js` extensions respectively.

### Configuration and secrets

External URLs, usernames, passwords, etc should be defined as flags.
Flags should be defined close to where they're used, but hoisted to
keep packages reusable where it makes sense. e.g. the
`external/atproto` package takes configuration, and the flags are
defined at the call site in `content`.

### Vertical feature slices

New code should prefer keeping code and resources together in their
feature domains: DB operations, CSS, business logic, etc, should
be in a single package. New HTTP handlers should be a thin wiring
layer that hands over to the feature package.

Shortcodes and their CSS should be defined within the relevant
feature slice where possible, rather than in `features/shortcodes/`
(which is only for truly cross-cutting shortcodes).

When doing this, DB operations should be placed in a `db.go` file
in the package. These operations should be as minimal as possible:
simple inserts or retrievals. There should be as little business
logic as possibl in the database files.

### Shortcode and asset registration

Packages that expose shortcodes should define a
`RegisterShortcodes(mgr *shortcodes.Manager)` function. Packages
that need to register static assets should define a
`RegisterAssets(mgr *assets.Manager)` function. These are
automatically discovered by the code generator (`cmd/generate`)
and written to `cmd/serve/register.go`. There is no need to
manually wire up new registrations — just add the function to
the package and run `go generate ./cmd/serve/`.

### Admin interface

The admin interface is only accessible by a single user. We do
not need to be concerned about parallel updates, authentication,
etc.

## Code standards

### General

- Use the `log/slog` package for logging when required
- This is a personal website, it doesn't need unit tests or
  public API documentation.
- Structs used for JSON encoding/decoding in a single function
  should be defined inline within that function, not as
  package-level types.

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
