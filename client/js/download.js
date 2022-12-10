upload.modules.addmodule({
    name: 'download',
    delkeys: {},
    // Dear santa, https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/template_strings
    template: '\
      <div class="modulecontent" id="dlarea">\
        <div class="topbar">\
        <h1 id="downloaded_filename"></h1>\
        <div class="viewswitcher">\
          <a class="btn" id="newupload" href="#">New Upload</a>\
        </div>\
        </div>\
        <div id="downloaddetails"></div>\
        <div id="btnarea">\
                <a class="btn" id="dlbtn" href="#">Download</a\
                ><a class="btn" id="inbrowserbtn" target="_blank" href="#">View In Browser</a\
                ><a class="btn" id="deletebtn" href="#">Delete</a\
                ><div class="right"><a class="btn" id="prevbtn" href="#">Prev</a\
                ><a class="btn" id="nextbtn" href="#">Next</a></div>\
        </div>\
      </div>\
    ',
    init: function () {
    },
    route: function (route, content) {
        if (content != 'noref') {
            return this
        }
    },
    render: function (view) {
        view.html(this.template)
        this._ = {}
        this._.view = view
        this._.detailsarea = view.find('#downloaddetails')
        this._.filename = view.find('#downloaded_filename')
        this._.btns = view.find('#btnarea')
        this._.deletebtn = view.find('#deletebtn')
        this._.dlbtn = view.find('#dlbtn')
        this._.nextbtn = view.find('#nextbtn')
        this._.prevbtn = view.find('#prevbtn')
        this._.viewbtn = view.find('#inbrowserbtn')
        this._.viewswitcher = view.find('.viewswitcher')
        this._.newupload = view.find('#newupload')
        this._.dlarea = view.find('#dlarea')
        this._.title = $('title')
        $('#footer').hide()
    },
    initroute: function (content, contentroot) {
        contentroot = contentroot ? contentroot : content
        this._.nextbtn.hide()
        this._.prevbtn.hide()
        if (contentroot.indexOf('&') > -1) {
          var which = 0
          var values = contentroot.split('&')
          var howmany = values.length
          if (content != contentroot) {
            which = parseInt(content) - 1
          }
          content = values[which]
          this._.nextbtn.attr('href', '#' + contentroot + '/' + (which + 2))
          this._.prevbtn.attr('href', '#' + contentroot + '/' + (which))
          if (!(which >= howmany - 1)) {
            this._.nextbtn.show()
          }
          if (!(which <= 0)) {
            this._.prevbtn.show()
          }
        }
        console.log(contentroot)
        delete this._['text']
        this._.filename.hide()
        this._.title.text("Up1")
        this._.btns.hide()
        this._.newupload.hide()
        this._.content = {}
        this._.content.main = this._.content.loading = $('<h1>').prop('id', 'downloadprogress').addClass('centertext centerable').text('Downloading')
        this._.detailsarea.empty().append(this._.content.main)
        this._.deletebtn.hide()
        upload.updown.download(content, this.progress.bind(this), this.downloaded.bind(this))
    },
    unrender: function () {
        this._.title.text('Up1')
        delete this['_']
    },

    downloaded: function (data) {
        this._.filename.text(data.header.name)
        this._.title.text(data.header.name + ' - Up1')

        var stored = this.delkeys[data.ident]

        if (!stored) {
            try {
                stored = localStorage.getItem('delete-' + data.ident)
            } catch (e) {
                console.log(e)
            }
        }

        if (stored && !isiframed()) {
            this._.deletebtn.show().prop('href', (upload.config.server ? upload.config.server : '') + 'del?delkey=' + stored + '&ident=' + data.ident)
        }

        this._.newupload.show()

        var decrypted = new Blob([data.decrypted], { type: data.header.mime })

        var safedecrypted = new Blob([decrypted], { type: data.header.mime })

        var url = URL.createObjectURL(decrypted)

        var safeurl = URL.createObjectURL(safedecrypted)

        this._.viewbtn.prop('href', safeurl).hide()
        this._.dlbtn.prop('href', url)
        this._.dlbtn.prop('download', data.header.name)

        delete this._['content']
        this._.detailsarea.empty()

        this._.viewbtn.show()

        $('<div>').addClass('preview').addClass('downloadexplain centerable centertext').text("Click the Download link in the bottom-left to download this file.").appendTo(this._.detailsarea)
        this._.filename.show()
        this._.btns.show()
    },
    progress: function (e) {
        if (e == 'decrypting') {
            this._.content.loading.text('Decrypting')
        } else if (e == 'error') {
          this._.content.loading.text('File not found or corrupt')
          this._.newupload.show()
        } else {
            var text = ''
            if (e.eventsource != 'encrypt') {
                text = 'Downloading'
            } else {
                text = 'Decrypting'
            }
            var percent = (e.loaded / e.total) * 100
            this._.content.loading.text(text + ' ' + Math.floor(percent) + '%')
        }
    }
})
