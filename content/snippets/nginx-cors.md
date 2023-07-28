---
title: Allow CORS requests from certain origins
group: Nginx
---

Instead of just sending `*` in the `Access-Control-Allow-Origin` header we can use a map to conditionally set it if the origin matches certain rules:

```nginx
map $http_origin $allow_origin {
    ~^https://.*?\.example.com$ $http_origin;
}
add_header Access-Control-Allow-Origin $allow_origin;
```