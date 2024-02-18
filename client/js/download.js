
import './loadencryption.js'
import '../deps/zepto.min.js'
import * as updown from './updown.js'
import * as config from '../config.js'

const _ = {}

function init (view, content) {
  _.view = view
  _.preview = view.find('#preview')
  _.loading = view.find('#downloadprogress')
  _.description = view.find('#description')
  _.daysuntilexpiry = view.find('#daysuntilexpiry')
  _.filename = view.find('#downloaded_filename')
  _.deletebtn = view.find('#deletebtn')
  _.savebtn = view.find('#savebtn')
  _.viewbtn = view.find('#inbrowserbtn')

  _.deletebtn.hide()
  _.savebtn.hide()
  updown.download(content, progress.bind(this), downloaded.bind(this))
}

// preview mime types
const previewassocations = {
  'image/tiff': 'image',
  'image/x-tiff': 'image',
  'image/bmp': 'image',
  'image/x-windows-bmp': 'image',
  'image/gif': 'image',
  'image/x-icon': 'image',
  'image/jpeg': 'image',
  'image/pjpeg': 'image',
  'image/png': 'image',
  'image/webp': 'image',
  'image/svg+xml': 'svg'
}

function getassociation(mime) {
  for (var key in previewassocations) {
      if (mime.startsWith(key)) {
          return previewassocations[key]
      }
  }
}

function downloaded (fileinfo, data) {
  const fi = JSON.parse(fileinfo)
  _.description.text(fi.Description)
  _.daysuntilexpiry.text(fi.DaysUntilExpiry)
  const fn = data.header.name || fi.Description + '.zip'
  _.filename.text(fn)
  _.loading.hide()

  if (fi.ViewerCanDelete) {
    _.deletebtn.show().click(function () {
      $.get('/del?ident=' + data.ident, function (res) {
        $('#topbar').hide()
        _.savebtn.hide()
        _.preview.hide()
        let msg = 'Deleted: ' + fi.Description
        if (res !== '') {
          const j = JSON.parse(res)
          msg = 'Error deleting: ' + j.error
        }
        $('.contentarea').addClass('bigcenterbutton').text(msg)
      })
    })
  }

  var association = getassociation(data.header.mime)
  const decrypted = new Blob([data.decrypted], { type: data.header.mime })
  const safedecrypted = new Blob([decrypted], { type: data.header.mime })
  const url = URL.createObjectURL(decrypted)
  const safeurl = URL.createObjectURL(safedecrypted)

  _.viewbtn.prop('href', safeurl)
  _.savebtn.show()
  _.savebtn.prop('href', url)
  _.savebtn.prop('download', fn)

  if (association == 'image' || association == 'svg') {
    $('<img>').appendTo(_.preview).prop('src', url)
  } else if (config.downloadautomatically) {
    _.savebtn.click()
    _.savebtn.text('File has been saved to download folder, click here to save it again!')
  }
}

function progress (e) {
  if (e === 'decrypting') {
    _.loading.text('Decrypting')
  } else if (e === 'error') {
    _.loading.text('File not found or corrupt')
    _.savebtn.hide()
    _.viewbtn.hide()
  } else {
    let text = ''
    if (e.eventsource !== 'encrypt') {
      text = 'Downloading'
    } else {
      text = 'Decrypting'
    }
    const percent = (e.loaded / e.total) * 100
    _.loading.text(text + ' ' + Math.floor(percent) + '%')
  }
}

(function () {
  const view = $('body')
  init(view, window.location.hash.substring(1))
}())
