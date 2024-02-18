importScripts('../deps/sjcl.min.js')
// can't get sjcl encryption worker (loadencryption) to work with es6 modules

function parametersfrombits (seed) {
  const out = sjcl.hash.sha512.hash(seed)
  return {
    seed,
    key: sjcl.bitArray.bitSlice(out, 0, 256),
    iv: sjcl.bitArray.bitSlice(out, 256, 384),
    ident: sjcl.bitArray.bitSlice(out, 384, 512)
  }
}

function parameters (seed) {
  if (typeof seed === 'string') {
    seed = sjcl.codec.base64url.toBits(seed)
  } else {
    seed = sjcl.codec.bytes.toBits(seed)
  }
  return parametersfrombits(seed)
}

function encrypt (file, seed, id) {
  try {
    const params = parameters(seed)
    const uarr = new Uint8Array(file)
    const before = sjcl.codec.bytes.toBits(uarr)
    const prp = new sjcl.cipher.aes(params.key)
    const after = sjcl.mode.ccm.encrypt(prp, before, params.iv)
    const afterarray = new Uint8Array(sjcl.codec.bytes.fromBits(after))
    postMessage({
      id,
      seed: sjcl.codec.base64url.fromBits(params.seed),
      ident: sjcl.codec.base64url.fromBits(params.ident),
      encrypted: new Blob([afterarray], { type: 'application/octet-stream' })
    })
  } catch(error) {
    console.log("encrypt error: ", error)
    postMessage({
      id,
      eventsource: 'encrypt',
      type: 'progress',
      error: 'encryption error, file too big?'
    })
  }

}

const fileheader = [
  85, 80, 49, 0
]

function decrypt (file, seed, id) {
  const params = parameters(seed)
  let uarr = new Uint8Array(file)

  // We support the servers jamming a header in to deter direct linking
  let hasheader = true
  for (let i = 0; i < fileheader.length; i++) {
    if (uarr[i] !== fileheader[i]) {
      hasheader = false
      break
    }
  }
  if (hasheader) {
    uarr = uarr.subarray(fileheader.length)
  }

  const before = sjcl.codec.bytes.toBits(uarr)
  const prp = new sjcl.cipher.aes(params.key)
  const after = sjcl.mode.ccm.decrypt(prp, before, params.iv)
  const afterarray = new Uint8Array(sjcl.codec.bytes.fromBits(after))

  // Parse the header, which is a null-terminated UTF-16 string containing JSON
  let header = ''
  const headerview = new DataView(afterarray.buffer)
  let i = 0
  for (; ; i++) {
    const num = headerview.getUint16(i * 2, false)
    if (num === 0) {
      break
    }
    header += String.fromCharCode(num)
  }
  header = JSON.parse(header)

  const data = new Blob([afterarray])
  postMessage({
    id,
    ident: sjcl.codec.base64url.fromBits(params.ident),
    header,
    decrypted: data.slice((i * 2) + 2, data.size, header.mime)
  })
}

function ident (seed, id) {
  const params = parameters(seed)
  postMessage({
    id,
    ident: sjcl.codec.base64url.fromBits(params.ident)
  })
}

function onprogress (id, progress) {
  postMessage({
    id,
    eventsource: 'encrypt',
    loaded: progress,
    total: 1,
    type: 'progress'
  })
}

self.onmessage = function (e) {
  const progress = onprogress.bind(undefined, e.data.id)
  sjcl.mode.ccm.listenProgress(progress)
  if (e.data.action === 'decrypt') {
    decrypt(e.data.data, e.data.seed, e.data.id)
  } else if (e.data.action === 'ident') {
    ident(e.data.seed, e.data.id)
  } else {
    sjcl.random.addEntropy(e.data.entropy, 2048, 'runtime')
    encrypt(e.data.data, e.data.seed, e.data.id)
  }
  sjcl.mode.ccm.unListenProgress(progress)
}
