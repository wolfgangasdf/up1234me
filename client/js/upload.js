
import "./loadencryption.js"
import * as config from "../config.js"
import * as updown from "./updown.js"
import "./shims.js"
import "../deps/zepto.min.js"

var _ = {}

function init() {
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

function dragleave(e) {
    e.preventDefault()
    e.stopPropagation()
    _.pastearea.removeClass('dragover')
}

function drop(e) {
    console.log("dropped!!!")
    e.preventDefault()
    _.pastearea.removeClass('dragover')
    if (e.dataTransfer.files.length > 0) {
        doupload(e.dataTransfer.files[0])
    }
}

function dragover(e) {
    e.preventDefault()
    _.pastearea.addClass('dragover')
}

function pickfile(e) {
    _.filepicker.click()
}

function pickerchange(e) {
    if (e.target.files.length > 0) {
        doupload(e.target.files[0])
        $(e.target).parents('form')[0].reset()
    }
}

function render(view) {
    _.view = view
    _.filepicker = view.find('#filepicker')
    _.pastearea = view.find('#pastearea')
    _.progress = {}
    _.progress.main = view.find('#uploadprogress')
    _.progress.type = view.find('#progresstype')
    _.progress.amount = view.find('#progressamount')
    _.progress.bg = view.find('#progressamountbg')
    _.beforeupload = view.find('#beforeupload')
    _.filename = view.find('#filename')
    _.description = view.find('#description')
    $('#footer').show()
}

function initroute() {
    focuspaste()
}

function unrender() {
    delete this['_']
}

function initpastecatcher() {
    var pastecatcher = $('<pre>').prop('id', 'pastecatcher')
    pastecatcher.prop('contenteditable', true)
    $('body').append(pastecatcher)
}
function vfocuspaste() {
    setTimeout(function () {
        pastecatcher.focus()
    }, 100)
}

function triggerfocuspaste(e) {
    if (e.which != 1) {
        return
    }

    if (e.target == document.body && _ && !_.pastearea.hasClass('hidden')) {
        e.preventDefault()
        focuspaste()
    }
}
function progress(e) {
    if (e.eventsource != 'encrypt') {
        _.progress.type.text('Uploading')
    } else {
        _.progress.type.text('Encrypting')
    }
    var percent = (e.loaded / e.total) * 100
    _.progress.bg.css('width', percent + '%')
    _.progress.amount.text(Math.floor(percent) + '%')
}
// WL this is now "prepare for upload" - rename later! TODO
function doupload(blob) {
    _.pastearea.addClass('hidden')

    _.beforeupload.removeClass('hidden')
    _.filename[0].innerHTML = blob.name
    _.description[0].value = blob.name
    _.blob = blob
}
function douploadreally() {
    _.progress.main.removeClass('hidden')
    _.progress.type.text('Encrypting')
    _.progress.bg.css('width', 0)
    updown.upload(_.blob, progress.bind(this), uploaded.bind(this))
}
function closepaste() {
  _.pastearea.removeClass('hidden')
  _.view.find('#uploadview').show()
  _.view.find('.viewswitcher').show()
}
function dopasteupload(data) {
    _.pastearea.addClass('hidden')
    _.view.find('#uploadview').hide()
    _.view.find('.viewswitcher').hide()
}

function uploaded(data, response) {
    // download.delkeys[data.ident] = response.delkey

    // try {
    //     localStorage.setItem('delete-' + data.ident, response.delkey)
    // } catch (e) {
    //     console.log(e)
    // }

    // TODO
    // if (window.location.hash == '#noref') {
    //     history.replaceState(undefined, undefined, '#' + data.seed)
    //     // route.setroute(download, undefined, data.seed)
    // } else {
    // }
    window.location = 'd/#' + data.seed

}
function pasted(e) {
    console.log("pasted!!!")
    if (!_ || _.pastearea.hasClass('hidden')) {
        return
    }

    var items = e.clipboardData.items

    if (typeof items != 'undefined' && items.length >= 1) {
        e.preventDefault()

        for (var i = 0; i < items.length; i++) {
            console.log("items[" + i + "]=" + items[i])
            var blob = items[i].getAsFile()
            if (blob) {
                doupload(blob)
                break
            }
        }

    }
}

(function () {
    console.log("2 "+document.getElementById("filename"))
    var view = $('.modulecontent.modulearea')
    console.log("asdf=" + view.find('#filename'))
    render(view)
    init()
}())



