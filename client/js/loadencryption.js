
window.crypt = {}
const crypto = window.crypto || window.msCrypto
const worker = new Worker('../js/encryption.js')


function getEntropy () {
  const entropy = new Uint32Array(256)
  crypto.getRandomValues(entropy)
  return entropy
}

function getSeed () {
  const seed = new Uint8Array(16)
  crypto.getRandomValues(seed)
  return seed
}

const promises = {}

function str2ab (str) {
  const buf = new ArrayBuffer(str.length * 2)
  const bufView = new DataView(buf)
  for (let i = 0, strLen = str.length; i < strLen; i++) {
    bufView.setUint16(i * 2, str.charCodeAt(i), false)
  }
  return buf
}

worker.onmessage = function (e) {
  if (e.data.type === 'progress') {
    promises[e.data.id].notify(e.data)
  } else {
    promises[e.data.id].resolve(e.data)
    delete promises[e.data.id]
  }
}

let counter = 0

function getpromise () {
  const promise = $.Deferred()
  const promiseid = counter
  counter += 1
  promise.id = promiseid
  promises[promiseid] = promise
  return promise
}

crypt.encrypt = function (file, name) {
  const header = JSON.stringify({
    mime: file.type,
    name: name || file.name
  })

  const zero = new Uint8Array([0, 0])

  const blob = new Blob([str2ab(header), zero, file])

  const promise = getpromise()

  const fr = new FileReader()

  fr.onload = function () {
    worker.postMessage({
      data: this.result,
      entropy: getEntropy(),
      seed: getSeed(),
      id: promise.id
    })
  }

  fr.readAsArrayBuffer(blob)

  return promise
}

crypt.ident = function (seed) {
  const promise = getpromise()

  worker.postMessage({
    seed,
    action: 'ident',
    id: promise.id
  })

  return promise
}

crypt.decrypt = function (file, seed) {
  const promise = getpromise()

  const fr = new FileReader()

  fr.onload = function () {
    worker.postMessage({
      data: this.result,
      action: 'decrypt',
      seed,
      id: promise.id
    })
  }

  fr.readAsArrayBuffer(file)

  return promise
}
