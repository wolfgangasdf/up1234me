import * as config from "../config.js"

var requestframe = {}

function init() {
    // We do this to try to hide the fragment from the referral in IE
    requestframe = document.createElement('iframe')
    requestframe.src = 'about:blank'
    requestframe.style.visibility = 'hidden'
    document.body.appendChild(requestframe)
}
function downloadfromident(seed, progress, done, ident) {
    var xhr = new requestframe.contentWindow.XMLHttpRequest()
    xhr.onload = downloaded.bind(this, seed, progress, done)
    xhr.open('GET', '/i/' + ident.ident)
    xhr.responseType = 'blob'
    xhr.onerror = onerror.bind(this, progress)
    xhr.addEventListener('progress', progress, false)
    xhr.send()
}

function onerror(progress) {
  progress('error')
}

function downloaded(seed, progress, done, response) {
    if (response.target.status != 200) {
      onerror(progress)
    } else {
      var fi = response.target.getResponseHeader("Fileinfo")
      if (!fi) {
        console.log("error reading file info")
        onerror(progress)
      } else {
        progress('decrypting')
        crypt.decrypt(response.target.response, seed).done(done.bind(this, fi))
      }
    }
}
function uploadencrypted(metadata, progress, done, data) {
    var formdata = new FormData()
    formdata.append('ident', data.ident)
    formdata.append('file', data.encrypted)
    formdata.append('description', metadata.description) // TODO in json?
    formdata.append('expirydays', metadata.expirydays)
    formdata.append('viewercandelete', metadata.viewercandelete)
    $.ajax({
        url: 'up',
        data: formdata,
        cache: false,
        processData: false,
        contentType: false,
        dataType: 'json',
        xhr: function () {
            var xhr = new XMLHttpRequest()
            xhr.upload.addEventListener('progress', progress, false)
            return xhr
        },
        type: 'POST'
    }).done(done.bind(undefined, data))
}
export function download(seed, progress, done) {
    crypt.ident(seed).done(downloadfromident.bind(this, seed, progress, done))
}
export function upload(blob, metadata, progress, done) {
  crypt.encrypt(blob).done(uploadencrypted.bind(this, metadata, progress, done)).progress(progress)
}

(function () {
  init()
}())