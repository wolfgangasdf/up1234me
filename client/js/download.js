import "./loadencryption.js"
import * as updown from "./updown.js"
import "./shims.js"
import "../deps/zepto.min.js"


var _ = {}
function route(route, content) {
    if (content != 'noref') {
        return this
    }
}

function render(view) {
    _ = {}
    _.view = view
    _.detailsarea = view.find('#downloaddetails')
    _.loading = view.find('#downloadprogress')
    _.description = view.find('#description')
    _.daysuntilexpiry = view.find('#daysuntilexpiry')
    _.viewercandelete = view.find('#viewercandelete')
    _.downloadcount = view.find('#downloadcount')
    _.filename = view.find('#downloaded_filename')
    _.btns = view.find('#btnarea')
    _.deletebtn = view.find('#deletebtn')
    _.dlbtn = view.find('#dlbtn')
    _.viewbtn = view.find('#inbrowserbtn')
    _.newupload = view.find('#newupload')
    _.dlarea = view.find('#dlarea')
    $(document).on('click', '#deletebtn', deleteupload.bind(this))
}
function initroute(content, contentroot) {
    contentroot = contentroot ? contentroot : content
    console.log("contentroot=", contentroot)
    delete _['text']
    _.filename.hide()
    _.btns.hide()
    _.newupload.hide()
    _.deletebtn.hide()

    updown.download(content, progress.bind(this), downloaded.bind(this))
}
function unrender() {
    delete this['_']
}

function downloaded(fileinfo, data) {
    const fi = JSON.parse(fileinfo);
    _.description.text("Description: " + fi.Description)
    _.daysuntilexpiry.text("Days until expiry: " + fi.DaysUntilExpiry)
    _.viewercandelete.text("Viewer can delete: " + fi.ViewerCanDelete)
    _.downloadcount.text("Downloads: " + fi.DownloadCount)
    const fn = data.header.name || fi.Description + ".zip" 
    _.filename.text(fn)
    _.loading.hide()

    if (fi.ViewerCanDelete) {
        _.deletebtn.show().prop('href', "http://" + window.location.host + '/del?ident=' + data.ident)
    }

    _.newupload.show()

    var decrypted = new Blob([data.decrypted], { type: data.header.mime })

    var safedecrypted = new Blob([decrypted], { type: data.header.mime })

    var url = URL.createObjectURL(decrypted)

    var safeurl = URL.createObjectURL(safedecrypted)

    _.viewbtn.prop('href', safeurl).hide()
    _.dlbtn.prop('href', url)
    _.dlbtn.prop('download', fn)

    delete _['content']
    _.detailsarea.empty()

    _.viewbtn.show()

    _.filename.show()
    _.btns.show()
}
function deleteupload() {
    console.log("delete upload: clear cache!")
    updown.cachedelete()
}

function progress(e) {
    if (e == 'decrypting') {
        _.loading.text('Decrypting')
    } else if (e == 'error') {
      _.loading.text('File not found or corrupt')
      _.newupload.show()
    } else {
        var text = ''
        if (e.eventsource != 'encrypt') {
            text = 'Downloading'
        } else {
            text = 'Decrypting'
        }
        var percent = (e.loaded / e.total) * 100
        _.loading.text(text + ' ' + Math.floor(percent) + '%')
    }
}

(function () {
  var view = $('.modulecontent.modulearea')
  render(view)

  console.log("location: " + window.location.hash)
  initroute(window.location.hash.substring(1))

}())
