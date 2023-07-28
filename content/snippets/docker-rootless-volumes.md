---
title: Volumes in a rootless container
group: Docker
---

To make docker volumes work nicely, you want the mount point owned by the user the application will be running as. One way to do this is create a directory in an earlier stage and copy it with `--chown`:

```dockerfile
FROM whatever AS build
RUN mkdir /data

FROM gcr.io/distroless/base:nonroot
COPY --from=build --chown=nonroot /data /data
VOLUME /data
```