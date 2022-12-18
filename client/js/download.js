import "./loadencryption.js"
import * as updown from "./updown.js"
import "./shims.js"
import "../deps/zepto.min.js"
import * as config from "../config.js"


var _ = {}

function render(view) {
    _.view = view
    _.detailsarea = view.find('#downloaddetails')
    _.loading = view.find('#downloadprogress')
    _.description = view.find('#description')
    _.daysuntilexpiry = view.find('#daysuntilexpiry')
    _.downloadcount = view.find('#downloadcount')
    _.filename = view.find('#downloaded_filename')
    _.btns = view.find('#btnarea')
    _.deletebtn = view.find('#deletebtn')
    _.dlbtn = view.find('#dlbtn')
    _.viewbtn = view.find('#inbrowserbtn')
    _.newupload = view.find('#newupload')
    _.dlarea = view.find('#dlarea')
}
function initroute(content, contentroot) {
    delete _['text']
    _.filename.hide()
    _.btns.hide()
    _.newupload.hide()
    _.deletebtn.hide()

    updown.download(content, progress.bind(this), downloaded.bind(this))
}

function downloaded(fileinfo, data) {
    const fi = JSON.parse(fileinfo);
    _.description.text(fi.Description)
    _.daysuntilexpiry.text(fi.DaysUntilExpiry)
    _.downloadcount.text(fi.DownloadCount)
    const fn = data.header.name || fi.Description + ".zip" 
    _.filename.text(fn)
    _.loading.hide()

    if (fi.ViewerCanDelete) {
        _.deletebtn.show().prop('href', "http://" + window.location.host + '/del?ident=' + data.ident)
    }

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

    if (_.calledByUpload) _.newupload.show()

    if (config.downloadautomatically && !_.calledByUpload) {
        _.dlbtn.click() 
        _.dlbtn.text("Download again")
    }
}

function progress(e) {
    if (e == 'decrypting') {
        _.loading.text('Decrypting')
    } else if (e == 'error') {
      _.loading.text('File not found or corrupt')
      _.newupload.show()
      _.dlbtn.hide()
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
    _.calledByUpload = false
    if (window.location.href.endsWith("&")) {
        console.log("called by upload!")
        window.location.replace(window.location.href.slice(0, -1))
        _.calledByUpload = true
    }

  var view = $('body')
  render(view)

  console.log("location: " + window.location.hash, " win: ", window, document['variab'])
  initroute(window.location.hash.substring(1))

}())
