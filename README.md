# Chameth.com

This repo contains the source for my personal website hosted at https://chameth.com/

![](./cmd/serve/assets/static/screenshot.png?raw=true)

The site is running a custom Go server found in `cmd/serve`.

## Admin login

The admin interface is served on the tailnet and requires an OpenID Connect
login. Configure it with these flags (or the equivalent env vars, e.g.
`OIDC_CLIENT_ID`):

| Flag                  | Default               | Purpose                                                  |
|-----------------------|-----------------------|----------------------------------------------------------|
| `-oidc-issuer`        | `https://id.929w.net` | OpenID Connect issuer                                    |
| `-oidc-client-id`     |                       | Client ID                                                |
| `-oidc-client-secret` |                       | Client secret                                            |
| `-oidc-session-key`   |                       | Signs session cookies; at least 32 characters            |
| `-no-login`           | `false`               | Serve the admin interface without a login, for local use |

The client's callback URL is `https://website-admin.yak-wall.ts.net/auth/callback`.
Sessions last until the next 05:00 Europe/London; changing the session key logs
everyone out.

The same client authenticates the site to the services it pulls data from,
using client-credentials tokens with each service's URL as the `resource`.
Temporal calls use `-temporal-resource` (default
`https://temporal.yak-wall.ts.net`) as theirs. Without a client ID and secret,
those requests are sent without a token.

## History

It was previously built using [11ty](https://www.11ty.dev/); you can see the
last revision [here](https://github.com/csmith/chameth.com/tree/4094173f88ebb4cf601e6dc3853e700cb00877fa).
Prior to that it used [Hugo](https://gohugo.io/); you can see the last revision
[here](https://github.com/csmith/chameth.com/tree/93f51f5ddf9af4f19320ccecdff85077a09f5559).
Prior to THAT it used [Jekyll](https://jekyllrb.com/); last revision
[here](https://github.com/csmith/chameth.com/tree/e57880cb9a118e1ddac60efe7db796d0007cc2d1).