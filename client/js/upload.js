
import './loadencryption.js'
import '../deps/zepto.min.js'
import * as cz from '../deps/client-zip.js'
import * as updown from './updown.js'

const _ = {}
_.blobs = []
_.filenames = []
_.filecount = 0

function init (view) {
  _.view = view
  _.filepicker = view.find('#filepicker')
  _.pastearea = view.find('#pastearea')
  _.progress = {}
  _.progress.main = view.find('#uploadprogress')
  _.progress.type = view.find('#progresstype')
  _.progress.amount = view.find('#progressamount')
  _.progress.bg = view.find('#progressamountbg')
  _.beforeupload = view.find('#beforeupload')
  _.filenamesdiv = view.find('#filenames')
  _.description = view.find('#description')
  _.expirydays = view.find('#expirydays')
  _.viewercandelete = view.find('#viewercandeleteup')
  _.uploaddone = view.find('#uploaddone')
  _.uploaddonea = view.find('#uploaddonea')
  _.uploaddoneclip = view.find('#uploaddoneclip')
  _.deletebtn = view.find('#deletebtn')
  $('#uploadreally').hide()
  _.progress.main.hide()
  _.uploaddone.hide()
  _.deletebtn.hide()

  $(document).on('change', '#filepicker', pickerchange.bind(this))
  $(document).on('click', '#pastearea', pickfile.bind(this))
  $(document).on('dragover', '#pastearea', dragover.bind(this))
  $(document).on('dragleave', '#pastearea', dragleave.bind(this))
  $(document).on('drop', '#pastearea', drop.bind(this))
  $(document).on('click', triggerfocuspaste.bind(this))
  initpastecatcher()
  $(document).on('paste', pasted.bind(this))
  $(document).on('click', '#uploadreally', douploadreally.bind(this))
}

function dragleave (e) {
  e.preventDefault()
  e.stopPropagation()
  _.pastearea.removeClass('dragover')
}

function drop (e) {
  e.preventDefault()
  _.pastearea.removeClass('dragover')
  douploadmultiple(e.dataTransfer.files)
}

function dragover (e) {
  e.preventDefault()
  _.pastearea.addClass('dragover')
}

function pickfile (e) {
  _.filepicker.click()
}

function pickerchange (e) {
  douploadmultiple(e.target.files)
  $(e.target).parents('form')[0].reset()
}

const pastecatcher = $('<pre>').prop('id', 'pastecatcher')

function initpastecatcher () {
  pastecatcher.prop('contenteditable', true)
  $('body').append(pastecatcher)
}

function focuspaste () {
  setTimeout(function () {
    pastecatcher.focus()
  }, 100)
}

function triggerfocuspaste (e) {
  if (e.which !== 1) {
    return
  }

  if (e.target === document.body && _ && !_.pastearea.hasClass('hidden')) {
    e.preventDefault()
    focuspaste()
  }
}

// this receives object from encryption.js, or ajax ProgressEvent, or error object from server
function progress (e) {
  let showpercent = false
  if (e.eventsource === 'encrypt') {
    _.progress.type.text('Encrypting')
    showpercent = true
  } else if (e.constructor.name === 'ProgressEvent') {
    _.progress.type.text('Uploading')
    showpercent = true
  } else if (e.error != null) { // error from server
    _.progress.type.text('Error')
    _.progress.amount.text(e.error)
  }
  if (showpercent) {
    const percent = (e.loaded / e.total) * 100
    _.progress.bg.css('width', percent + '%')
    _.progress.amount.text(Math.floor(percent) + '%')
  }
}

function douploadmultiple (files) {
  for (let i = 0; i < files.length; i++) {
    doupload(files[i])
  }
}

// this adds one file to upload
function doupload (blob) {
  $('#uploadreally').show()
  $('<div>').text(blob.name).appendTo(_.filenamesdiv[0])
  _.beforeupload.removeClass('hidden')
  if (_.filecount === 0) {
    _.description.val(blob.name)
  } else if (_.filenames[0] === _.description.val()) { // remove extension from description if description unchanged
    const s = _.description.val().split('.').slice(0, -1).join('.')
    _.description.val(s === '' ? _.description.val() : s)
  }
  _.blobs.push(blob)
  _.filenames.push(blob.name)
  _.filecount++
}

// called if "upload" botton clicked
function douploadreally () {
  _.progress.main.show()
  _.progress.type.text('Encrypting')
  _.progress.bg.css('width', 0)
  _.metadata = {
    description: _.description.val(),
    expirydays: _.expirydays.val() === '' ? _.expirydays.attr('placeholder') : _.expirydays.val(),
    viewercandelete: _.viewercandelete.attr('checked') || false
  }
  if (_.filecount === 1) {
    updown.upload(_.blobs[0], _.metadata, progress.bind(this), uploaded.bind(this))
  } else { // zip
    const compressed = cz.downloadZip(_.blobs).blob()
    compressed.then(function (b) {
      updown.upload(b, _.metadata, progress.bind(this), uploaded.bind(this))
    })
  }
}

function uploaded (data, response) {
  console.log('uploaded: response=', response, ' data=', data)
  if (!response) {
    _.progress.type.text('Error')
    _.progress.amount.text('response null')
  } else if (response.code) {
    _.progress.type.text('Error')
    _.progress.amount.text(response.error)
  } else {
    _.progress.main.hide()
    $('#uploadarea').hide()
    $('#uploadreally').hide()
    _.deletebtn.show().click(function () {
      $.get('/del?ident=' + data.ident, function (res) {
        if (res !== '') {
          const j = JSON.parse(res)
          $('#uploaddone').text('Error deleting: ' + j.error)
        } else {
          window.location.reload()
        }
      })
    })

    const secreturl = window.location.protocol + '//' + window.location.host + '/d/#' + data.seed
    _.uploaddone.show()
    _.uploaddonea[0].href = secreturl
    _.uploaddonea[0].textContent = secreturl

    navigator.clipboard.writeText(secreturl).then(function () { // need to wait until finished!
      console.log('Copying to clipboard was successful!')
    }, function (err) {
      console.error('Could not copy text to clipboard: ', err)
      _.uploaddoneclip.hide()
    })
  }
}
function pasted (e) {
  if (!_ || _.pastearea.hasClass('hidden')) {
    return
  }
  const items = e.clipboardData.items
  if (typeof items !== 'undefined' && items.length >= 1) {
    e.preventDefault()
    for (let i = 0; i < items.length; i++) {
      const blob = items[i].getAsFile()
      if (blob) {
        doupload(blob)
      }
    }
  }
}

(function () {
  if (window.location.href.endsWith('#')) {
    window.location.replace(window.location.href.slice(0, -1))
  }
  const view = $('body')
  init(view)
}())
