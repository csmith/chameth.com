---
title: Redirect to canonical domain
group: Nginx
---

Instead of complex rules to redirect certain requests, just add a separate server block:

```nginx
server {
    server_name www.example.com;
    return 301 $scheme://example.com$request_uri;
}
```

Or add a default server block to catch _everything_ not explicitly dealt with:

```nginx
server {
    server_name _;
    listen 80 default_server;
    return 301 $scheme://example.com$request_uri;
}
```