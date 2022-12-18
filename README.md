Up1234me: A Client-side Encrypted File Host based on [Up1](https://github.com/Upload/Up1)

With the added changes it is my personal dropbox for sensitive files.

===
Changes compared to Up1:
* http basic authentication for upload and admin
* it's for generic files (not images) which are now previewed in the browser
* files can be automatically downloaded if link opened
* multiple-file uploads, which are zipped
* simpler js, html, and css code (deparate html, ES6 modules)
* upload expiry
* uploads can be deleted by viewer if allowed
* unencrypted metadata: description, expiry, viewercandelete

===
For a quick start, simply move `server.conf.example` to `server.conf`.

`listen` is an `address:port`-formatted string, where either one are optional. Some examples include `":9000"` to listen on any interface, port 9000; `"127.0.0.1"` to listen on localhost port 80; `"1.1.1.1:8080"` to listen on 1.1.1.1 port 8080; or even `""` to listen on any interface, port 80.

`maximum_file_size` is the largest file, in bytes, that's allowed to be uploaded to the server. The default here is a decimal 50MB.

For the web application configuration, a `config.js.example` file is provided, copy over...

=== http auth ===
`cd server`

`htpasswd -c server.htpasswd <username>`

=== build ===

`go build && ./up1234me`

=== used ===
* [client-zip](https://github.com/Touffy/client-zip) via https://cdn.jsdelivr.net/npm/client-zip/index.js
* 

How it works
---

See Up1 for a documentation of the encryption logic.

Additionally, the repository copy of SJCL comes from the source at https://github.com/bitwiseshiftleft/sjcl, commit `fb1ba931a46d41a7c238717492b66201b2995840` (Version 1.0.3), built with the command line `./configure --without-all --with-aes --with-sha512 --with-codecBytes --with-random --with-codecBase64 --with-ccm`, and compressed using Closure Compiler. If all goes well, a self-built copy should match up byte-for-byte to the contents of `static/deps/sjcl.min.js`.

The server-side is a Go server which uses no dependencies outside of the standard library. The only cryptography it uses is for generating deletion keys, using HMAC and SHA256 in the built-in `crypto/hmac` and `crypto/sha256` packages, respectively.



License
---

Like Up1: MIT.
