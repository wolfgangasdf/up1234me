Up1234me: A client-side encrypted file host based on [Up1](https://github.com/Upload/Up1). It is my personal dropbox for sensitive files.

Changes compared to Up1
===
* http basic authentication for upload and admin
* it's for generic files which are now previewed in the browser
* files can be automatically downloaded if link opened
* multiple-file uploads, which are zipped
* simpler js, html, and css code (separate html)
* upload expiry, runs every day
* uploads can be deleted by viewer if uploader has allowed it
* unencrypted metadata: description, expiry, viewercandelete
* copies url to clipboard after upload
* a basic admin interface (not finished)
* single binary thanks to go-bindata


Getting started
===
* download an executable of `up1234me` or build yourselves (below)
* create `server.conf` based on [server.conf.example](server.conf.example)
  * create folder "i" where the files are stored with access for webapp user
* create `config.js` based on [config.js.example](config.js.example)
* create http basic auth file in the `server` folder: `htpasswd -c server.htpasswd <username>`
* run (as webapp) `./up1234me` or `./up1234me-linux-amd64`
* use a reverse proxy for https!


Build
===

go-bindata
---
This is used to make a single executable file containing the static files (html, js etc).
```
go get -u github.com/go-bindata/go-bindata/v3/... 
go-bindata client/...        # put static files into bindata.go
```
For debugging use
````
go-bindata -debug client/... # development: use normal files via bindata.go
````

cross-build
---
```
GOOS=linux GOARCH=amd64 go build -o up1234me-linux-amd64
```

or build and run
---
```
go build && ./up1234me`
```

Used libraries
===
* [Up1](https://github.com/Upload/Up1) Everything is based on Up1, but there are too many changes for a fork. Based on commit [90c525a](https://github.com/Upload/Up1/commit/90c525a05db43c1063b02dd6164bf645bd569c81).
* [client-zip](https://github.com/Touffy/client-zip) via https://cdn.jsdelivr.net/npm/client-zip/index.js
* [go-bindata](https://github.com/go-bindata/go-bindata)
* [sjcl](https://github.com/bitwiseshiftleft/sjcl) for crypto, js from Up1
* [zeptojs](http://zeptojs.com/) as jquery
* [go-http-auth](github.com/abbot/go-http-auth) for http basic auth


License
===

Like Up1: MIT.
