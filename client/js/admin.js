
import * as config from "../config.js"
import "../deps/zepto.min.js"

var _ = {}
_.files = []

function init() {
    $(document).on('click', '#deleteallbefore', deleteallbefore.bind(this))
    updateinfo()
    updatefilelist(0)
}

function render(view) {
  _.view = view
  _.totalfilecount = view.find('#totalfilecount')
  _.totalsize = view.find('#totalsize')
  _.filelist = view.find('#filelist')
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
          _.filelist.append($("<tr>")
            .append($("<th>").text("Description"))
            .append($("<th>").text("FileDate"))
            .append($("<th>").text("FileSize"))
            .append($("<th>").text("Expirydays"))
            .append($("<th>").text("DaysUntilExpiry"))
            .append($("<th>").text("Viewercandelete"))
            .append($("<th>").text("Downloadcount"))
          )
    for (const fii in data.FileList) {
            let fi = data.FileList[fii]
            _.filelist.append($("<tr>")
                .append($("<td>").text(fi.Description))
                .append($("<td>").text(fi.FileDate))
                .append($("<td>").text(fi.FileSize))
                .append($("<td>").text(fi.Expirydays))
                .append($("<td>").text(fi.DaysUntilExpiry))
                .append($("<td>").text(fi.Viewercandelete))
                .append($("<td>").text(fi.Downloadcount))
            )
        } 
      }).catch(function(err) {
        console.log('Fetch Error: ', err);
      });
}

function deleteallbefore() {

}

(function () {
    var view = $('body')
    render(view)
    init()
}())



