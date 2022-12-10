import * as config from "../config.js"

var cached = {}
var cached_seed = {}
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
    xhr.open('GET', (config.server ? config.server : '') + '/i/' + ident.ident)
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
      cache(seed, response.target.response)
      progress('decrypting')
      crypt.decrypt(response.target.response, seed).done(done)
    }
}
function encrypted(progress, done, data) {
    var formdata = new FormData()
    formdata.append('api_key', config.api_key)
    formdata.append('ident', data.ident)
    formdata.append('file', data.encrypted)
    $.ajax({
        url: (config.server ? config.server : '') + 'up',
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
function cache(seed, data) {
  cached = data
  cached_seed = seed
}
function cacheresult(data) {
  cache(data.seed, data.encrypted)
}
export function download(seed, progress, done) {
    if (cached_seed == seed) {
      progress('decrypting')
      crypt.decrypt(cached, seed).done(done).progress(progress)
    } else {
      crypt.ident(seed).done(downloadfromident.bind(this, seed, progress, done))
    }
}
export function upload(blob, progress, done) {
  crypt.encrypt(blob).done(encrypted.bind(this, progress, done)).done(cacheresult.bind(this)).progress(progress)
}

(function () {
  init()
}())