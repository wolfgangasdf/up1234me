
import * as config from "../config.js"
import "../deps/zepto.min.js"

var _ = {}
_.files = []

function init() {
    $(document).on('click', '#deleteallbefore', deleteallbefore.bind(this))
    updateinfo()
    updatefilelist(0)
}

function updateinfo() {
    fetch("/admin/get_info").then(function(response) {
        return response.json();
      }).then(function(data) {
        _.totalfilecount.text(data.Totalfilecount)
        _.totalsize.text(data.Totalsize)
      }).catch(function(err) {
        console.log('Fetch Error: ', err);
      });
}

function updatefilelist(pageindex) {
    let formData = new FormData();
    formData.append('startindex', 123)
    fetch("/admin/get_files",
        {
            body: formData,
            method: "POST"
        }
    ).then(function(response) {
        return response.json();
      }).then(function(data) {
        for (const fii in data.FileList) {
            let fi = data.FileList[fii]
            $("<DIV>").text(fi.Description + " " + fi.Filename).appendTo(_.filelist)
        } 
      }).catch(function(err) {
        console.log('Fetch Error: ', err);
      });

}

function deleteallbefore() {

}

function render(view) {
    _.view = view
    _.totalfilecount = view.find('#totalfilecount')
    _.totalsize = view.find('#totalsize')
    _.filelist = view.find('#filelist')
}


(function () {
    var view = $('body')
    render(view)
    init()
}())



