---
title: Redirect types
group: HTTP
---

| Code  | Meaning            | Permanent? | Changes to a GET? |
|-------|--------------------|------------|-------------------|
| `301` | Moved Permanently  | Yes        | Depends           |
| `302` | Found              | No         | Depends           |
| `303` | See Other          | No         | Yes               |
| `307` | Temporary Redirect | No         | No                |
| `308` | Permanent Redirect | Yes        | No                |