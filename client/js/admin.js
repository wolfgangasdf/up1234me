
import '../deps/zepto.min.js'

const _ = {}

function init () {
  updatefilelist(0)
}

function render (view) {
  _.view = view
  _.totalfilecount = view.find('#totalfilecount')
  _.totalsize = view.find('#totalsize')
  _.filelist = view.find('#filelist')
}

function deletefile (ident) {
  fetch('/admin/delete_file?' + ident).then(function () {
    updatefilelist()
  }).catch(function (err) {
    console.log('Delete Error: ', err)
  })
}

// https://stackoverflow.com/a/18650828
function formatBytes(a,b=2){if(!+a)return'0 Bytes';const c=0>b?0:b,d=Math.floor(Math.log(a)/Math.log(1024));return`${parseFloat((a/Math.pow(1024,d)).toFixed(c))} ${['Bytes','KB','MB','GB','TB','PB','EB','ZB','YB'][d]}`}

function updatefilelist () {
  const formData = new FormData()
  formData.append('startindex', 123)
  fetch('/admin/get_files',
    {
      body: formData,
      method: 'POST'
    }
  ).then(function (response) {
    return response.json()
  }).then(function (data) {
    _.totalfilecount.text(data.TotalFiles)
    _.totalsize.text(formatBytes(data.TotalSize))
    _.filelist.empty()
    _.filelist.append($('<tr>')
      .append($('<th>').text('Del'))
      .append($('<th>').text('Description'))
      .append($('<th>').text('Date'))
      .append($('<th>').text('Size'))
      .append($('<th>').text('Expiry (d)'))
      .append($('<th>').text('Days left'))
      .append($('<th>').text('Can delete'))
      .append($('<th>').text('DL #'))
    )
    for (const fii in data.FileList) {
      const fi = data.FileList[fii]
      const b = $('<button>').text('X').click(function () { deletefile(fi.FileName) })
      _.filelist.append($('<tr>')
        .append($('<td>').append(b))
        .append($('<td>').text(fi.Saved.Description))
        .append($('<td>').text(fi.FileDate))
        .append($('<td>').text(formatBytes(fi.FileSize, 0)))
        .append($('<td>').text(fi.Saved.Expirydays))
        .append($('<td>').text(fi.DaysUntilExpiry))
        .append($('<td>').text(fi.Saved.Viewercandelete))
        .append($('<td>').text(fi.Saved.Downloadcount))
      )
    }
  }).catch(function (err) {
    console.log('Fetch Error: ', err)
  })
}

(function () {
  const view = $('body')
  render(view)
  init()
}())
