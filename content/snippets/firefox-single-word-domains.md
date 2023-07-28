---
title: Stop searching for single word domains
group: Firefox
---

Firefox strongly dislikes connecting to single word domains (e.g. "go/foo"), it much prefers searching for them even if you've visited the domain before. To fix it go to `about:config` and add an entry under `browser.fixup.domainwhitelist`, e.g. `browser.fixup.domainwhitelist.go=true`