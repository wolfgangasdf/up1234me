Up1234me: A client-side encrypted file host based on [Up1](https://github.com/Upload/Up1). Basically, the random private key after `#` in the URL is used for encryption and decryption in the browser (javascript), this key is never sent to the server. See [Up1](https://github.com/Upload/Up1) for details.


Changes compared to Up1
===
* it's for generic files, not mainly for pictures
* http basic authentication for upload and admin (mostly for single upload user)
* simpler html & js, I didn't get it :-), server golang only
* unencrypted metadata: description, expiry, viewercandelete
* a very basic administrator interface (list, delete)
* single server binary thanks to go-bindata
* upload:
  * multiple-file uploads, which are zipped in the browser
  * secret url is copied to clipboard after upload
  * upload expiry, runs every day
  * limit total storage size
* download:
  * files can be automatically downloaded if link opened
  * only images+svg are previewed in browser (automatic download is disabled for them)
  * uploads can be deleted by viewer if uploader has allowed it


Getting started
===
* download an [executable](https://github.com/wolfgangasdf/up1234me/releases) of `up1234me` or build yourselves (below), then do in the same folder:
* create `config.js` based on [config.js.example](config.js.example)
* create http basic auth file in the `server` folder: `htpasswd -c server.htpasswd <username>`
* create `server.conf` based on [server.conf.example](server.conf.example) and create the writable folder "storage_path"
* run (as webapp user) for example `./up1234me-linux-amd64`
* use a reverse proxy for https
* go to `/upload` and drag'n'drop, click or paste files to add to upload list. Click "Upload!" to upload.


Build
===

go-bindata
---
This is used to make a single executable file also containing the static files under `client/` (html, js, css etc.).
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
go build && ./up1234me
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
