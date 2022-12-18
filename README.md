Up1234me: A Client-side Encrypted File Host based on [Up1](https://github.com/Upload/Up1). It is my personal dropbox for sensitive files.

Changes compared to Up1
===
* http basic authentication for upload and admin
* it's for generic files (not images) which are now previewed in the browser
* files can be automatically downloaded if link opened
* multiple-file uploads, which are zipped
* simpler js, html, and css code (separate html)
* upload expiry, runs every day
* uploads can be deleted by viewer if uploader has allowed it
* unencrypted metadata: description, expiry, viewercandelete
* copies url to clipboard after upload
* a basic admin interface (not finished)
* single binary thans to go-bindata


Getting started
===
* copy `server.conf.example` to `server.conf` and adapt
* copy `config.js.example` to `config.js` and adapt
* download release or compile (below)
* create http basic auth file in the `server` folder: `htpasswd -c server.htpasswd <username>`
* make reverse SSL proxy
* run `./up1234me`


Build
===

Prepare static bindata (to have single executable)
---
```
go get -u github.com/go-bindata/go-bindata/v3/... 
# one of:
go-bindata        client/...      # put static files into bindata.go
go-bindata -debug client/... # development: use normal files via bindata.go
```

Cross-build
---
`GOOS=linux GOARCH=amd64 go build -o up1234me-linux-amd64`

Build and run
---
`go build && ./up1234me`


Used libraries
===
* [client-zip](https://github.com/Touffy/client-zip) via https://cdn.jsdelivr.net/npm/client-zip/index.js
* [Up1](https://github.com/Upload/Up1) Everything is based on this, too many changes to fork it. Commit version [90c525a](https://github.com/Upload/Up1/commit/90c525a05db43c1063b02dd6164bf645bd569c81).


License
===

Like Up1: MIT.
