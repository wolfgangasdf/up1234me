
import "../deps/zepto.min.js"
import * as updown from "./updown.js"



export var delkeys = {}
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
    _.filename = view.find('#downloaded_filename')
    _.btns = view.find('#btnarea')
    _.deletebtn = view.find('#deletebtn')
    _.dlbtn = view.find('#dlbtn')
    _.nextbtn = view.find('#nextbtn')
    _.prevbtn = view.find('#prevbtn')
    _.viewbtn = view.find('#inbrowserbtn')
    _.viewswitcher = view.find('.viewswitcher')
    _.newupload = view.find('#newupload')
    _.dlarea = view.find('#dlarea')
    _.title = $('title')
    $('#footer').hide()
}
function initroute(content, contentroot) {
    contentroot = contentroot ? contentroot : content
    _.nextbtn.hide()
    _.prevbtn.hide()
    if (contentroot.indexOf('&') > -1) {
      var which = 0
      var values = contentroot.split('&')
      var howmany = values.length
      if (content != contentroot) {
        which = parseInt(content) - 1
      }
      content = values[which]
      _.nextbtn.attr('href', '#' + contentroot + '/' + (which + 2))
      _.prevbtn.attr('href', '#' + contentroot + '/' + (which))
      if (!(which >= howmany - 1)) {
        _.nextbtn.show()
      }
      if (!(which <= 0)) {
        _.prevbtn.show()
      }
    }
    console.log(contentroot)
    delete _['text']
    _.filename.hide()
    _.title.text("Up1")
    _.btns.hide()
    _.newupload.hide()
    _.content = {}
    _.content.main = _.content.loading = $('<h1>').prop('id', 'downloadprogress').addClass('centertext centerable').text('Downloading')
    _.detailsarea.empty().append(_.content.main)
    _.deletebtn.hide()
    updown.download(content, progress.bind(this), downloaded.bind(this))
}
function unrender() {
    _.title.text('Up1')
    delete this['_']
}

function downloaded(data) {
    _.filename.text(data.header.name)
    _.title.text(data.header.name + ' - Up1')

    var stored = delkeys[data.ident]

    if (!stored) {
        try {
            stored = localStorage.getItem('delete-' + data.ident)
        } catch (e) {
            console.log(e)
        }
    }

    if (stored && !isiframed()) {
        _.deletebtn.show().prop('href', (upload.config.server ? upload.config.server : '') + 'del?delkey=' + stored + '&ident=' + data.ident)
    }

    _.newupload.show()

    var decrypted = new Blob([data.decrypted], { type: data.header.mime })

    var safedecrypted = new Blob([decrypted], { type: data.header.mime })

    var url = URL.createObjectURL(decrypted)

    var safeurl = URL.createObjectURL(safedecrypted)

    _.viewbtn.prop('href', safeurl).hide()
    _.dlbtn.prop('href', url)
    _.dlbtn.prop('download', data.header.name)

    delete _['content']
    _.detailsarea.empty()

    _.viewbtn.show()

    $('<div>').addClass('preview').addClass('downloadexplain centerable centertext').text("Click the Download link in the bottom-left to download this file.").appendTo(_.detailsarea)
    _.filename.show()
    _.btns.show()
}
function progress(e) {
    if (e == 'decrypting') {
        _.content.loading.text('Decrypting')
    } else if (e == 'error') {
      _.content.loading.text('File not found or corrupt')
      _.newupload.show()
    } else {
        var text = ''
        if (e.eventsource != 'encrypt') {
            text = 'Downloading'
        } else {
            text = 'Decrypting'
        }
        var percent = (e.loaded / e.total) * 100
        _.content.loading.text(text + ' ' + Math.floor(percent) + '%')
    }
}

(function () {
  var view = $('.modulecontent.modulearea')
  render(view)
  initroute(view)
}())
