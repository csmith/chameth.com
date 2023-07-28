---
title: Serve WebP images
group: Nginx
---

To manually convert images using `cwebp`:

```shell
cwebp -m 6 -mt -o "$file.webp" -- "$file"
```

To make nginx serve `.webp` files if the browser sends an accept header:

```nginx
http {
    map $http_accept $webp_suffix {
        "~*webp"  ".webp";
    }

    server {
        location ~ \.(png|jpe?g)$ {
            try_files $uri$webp_suffix $uri =404;
            expires 1y;
        }
    }
}
```