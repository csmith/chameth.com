---
title: Alert codes
group: TLS
---

| Code  | Meaning                                               | 
|-------|-------------------------------------------------------|
| `0`   | Close notify                                          |
| `10`  | Unexpected message                                    |
| `20`  | Bad record MAC                                        |
| `21`  | Decryption failed                                     |
| `22`  | Record overflow                                       |
| `30`  | Decompression failed                                  |
| `40`  | Handshake failed                                      |
| `41`  | No certificate (SSL 3.0)                              |
| `42`  | Certificate is bad                                    |
| `43`  | Certificate is not supported                          |
| `44`  | Certificate was revoked                               |
| `45`  | Certificate is expired                                |
| `46`  | Unknown certificate                                   |
| `47`  | Illegal parameter                                     |
| `48`  | CA is unknown                                         |
| `49`  | Access was denied                                     |
| `50`  | Decode error                                          |
| `51`  | Decrypt error                                         |
| `60`  | Export restriction                                    |
| `70`  | Error in protocol version                             |
| `71`  | Insufficient security                                 |
| `80`  | Internal error                                        |
| `86`  | Inappropriate fallback                                |
| `90`  | User canceled                                         |
| `100` | No renegotiation is allowed                           |
| `109` | An extension was expected but was not seen            |
| `110` | An unsupported extension was sent                     |
| `111` | Could not retrieve the specified certificate          |
| `112` | The server name sent was not recognized               |
| `115` | The SRP/PSK username is missing or not known          |
| `116` | Certificate is required                               |
| `120` | No supported application protocol could be negotiated |