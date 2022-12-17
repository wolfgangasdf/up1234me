
import "./loadencryption.js"
import * as config from "../config.js"
import * as updown from "./updown.js"
import "./shims.js"
import "../deps/zepto.min.js"
import * as cz from '../deps/client-zip.js';


var _ = {}
_.blobs = []
_.filenames = []
_.filecount = 0

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
    _.filenames = view.find('#filenames')
    _.description = view.find('#description')
    _.expirydays = view.find('#expirydays')
    _.viewercandelete = view.find('#viewercandeleteup')
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
// this is called if file dropped and prepares upload.
function doupload(blob) {
    // _.pastearea.addClass('hidden')
    $('<h2>').text(blob.name).appendTo(_.filenames[0])
    _.beforeupload.removeClass('hidden')
    if (_.filecount == 0) _.description.val(blob.name)
    _.blobs.push(blob)
    _.filenames.push(blob.name)
    _.filecount++
}

// called if "upload" botton clicked
function douploadreally() {
    _.progress.main.removeClass('hidden')
    _.progress.type.text('Encrypting')
    _.progress.bg.css('width', 0)
    _.metadata = {
        description: _.description.val(),
        expirydays: _.expirydays.val() == '' ? _.expirydays.attr('placeholder') : _.expirydays.val(),
        viewercandelete: _.viewercandelete.attr("checked") || false
    }
    if (_.filecount == 1) {
        updown.upload(_.blobs[0], _.metadata, progress.bind(this), uploaded.bind(this))
    } else { // zip
        const compressed = cz.downloadZip(_.blobs).blob()
        compressed.then(function (b) { // TODO this should be with progress in updown...
            updown.upload(b, _.metadata, progress.bind(this), uploaded.bind(this))
        })
    }
}

function uploaded(data, response) {
    console.log("uploaded: response=", response, " data=", data)
    if (response.code) {
        _.progress.type.text("Error")
        _.progress.amount.text(response.error)
    } else {
        window.location = 'd/#' + data.seed
    }
    // TODO
    // if (window.location.hash == '#noref') {
    //     history.replaceState(undefined, undefined, '#' + data.seed)
    //     // route.setroute(download, undefined, data.seed)
    // } else {
    // }

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
    var view = $('.modulecontent.modulearea')
    render(view)
    init()
}())



