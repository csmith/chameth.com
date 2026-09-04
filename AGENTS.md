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

#### Cached data shortcodes

Shortcodes that derive their output from data fetched from an external
source (e.g. a personal service or a social network) should register their
retrieve and render functions via `mgr.RegisterData(name, version, retrieve,
render)`. The framework
caches the retrieved response in the `shortcode_data` table, keyed by
`(shortcode, version, sha256(json.Marshal(args)))`, so the site stores
just the service's answer rather than the upstream data model. Usage in
content is identical to a regular shortcode.

The retrieve function returns `Result[T]`; the render function receives the
cached `T`. The data type must round-trip through JSON.

`Result.RefreshAt` is the exact time the data should next be refreshed;
a zero value freezes it (`next_refresh_at` is NULL). `NextRefresh` helps
compute a time from a refresh interval and optional cutoff, clamping the
last scheduled refresh to the cutoff and returning zero once it has passed.

Rendering never blocks on refresh: with data present the shortcode renders
from cache, due rows are refreshed by a background goroutine every 5 minutes,
and only the first-ever render of a key retrieves synchronously. A failed
retrieval is negative-cached: the row stores no data, further renders fail
fast with the standard shortcode error (no network calls), and the refresher
retries it every 5 minutes. Existing data is never overwritten by a failed
fetch.

Define the version as a named constant in the shortcode's `shortcode.go` file.
If the data shape or refresh semantics change, bump it so existing rows are
ignored. Rows for old/unregistered `(name, version)` pairs are intentionally
left in place.

### CSS

Global frontend styles are in `assets/stylesheet`. CSS must NEVER be
inlined in HTML in the frontend. Selectors should be nested where
possible/appropriate. Public CSS and JS files that should be bundled
must use the `.public.css` and `.public.js` extensions respectively.

### Configuration and secrets

Usernames, passwords, and other secrets should be defined as flags.
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

### Registration functions

Packages can define registration functions that are automatically
discovered by the code generator (`cmd/generate`) and wired into
`cmd/serve/register.go`. There is no need to manually wire up new
registrations — just add the function to the package and run
`go generate ./cmd/serve/`.

- `RegisterShortcodes(mgr *shortcodes.Manager)` — for shortcode registration
- `RegisterAssets(mgr *assets.Manager)` — for static asset registration
- `RegisterRoutes(mux *http.ServeMux)` — for HTTP route registration
- `RegisterGoroutine(...) func()` — for background goroutines, returns a function to launch

All registration functions can optionally accept parameters whose types match
fields on the `site` struct (e.g. `context.Context`, `*assets.Manager`,
`*tsnet.Server`). The generator inspects the function signature and passes the
matching arguments automatically.

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
- Return type structs should be defined in the `model.go`
  file not inline in other database files.
- Do not create "down" migrations, we only roll forwards.

## Commands

To build the app and test it compiles:

```
make build
```

To run verification (build + vet + fix + staticcheck + fmt):

```
make verify
```

To query the database:

```
docker compose exec -T database psql -U postgres -c "SELECT * FROM films;"
```

## Data preservation rules

- Do NOT touch the dev DB or tailscale folders (`.postgres` and `tsdata`)
- Do NOT attempt to run service: the user will do manual verification
- Do NOT manually modify the database
