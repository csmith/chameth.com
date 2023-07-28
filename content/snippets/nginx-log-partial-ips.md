---
title: Log only partial IP addresses
group: Nginx
---

This assumes nginx is behind a proxy that sends a trusted X-Forwarded-For header, but the same can easily be done with the remote IP directly.

```nginx
server {
    map $http_x_forwarded_for $forwarded_anon {
        ~(?P<ip>\d+\.\d+\.\d+)\.    $ip.0;
        ~(?P<ip>[^:]+:[^:]+):       $ip::;
        default                     0.0.0.0;
    }

    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$forwarded_anon"';
}
```