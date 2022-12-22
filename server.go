package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	auth "github.com/abbot/go-http-auth"
)

type Config struct {
	Listen         string `json:"listen"`
	MaxFileSize    int64  `json:"maximum_file_size"`
	MaxStorageSize int64  `json:"maximum_storage_size"`
	StoragePath    string `json:"storage_path"`
}

var config Config

type Metadata struct {
	Description     string
	Expirydays      int
	DaysUntilExpiry int `json:",omitempty"` // calculated at load
	Viewercandelete bool
	Downloadcount   int
	FileDate        time.Time `json:",omitempty"`
	FileSize        int64     `json:",omitempty"`
	FileName        string    `json:",omitempty"`
}

func (d *Metadata) MarshalJSON() ([]byte, error) { // format date https://stackoverflow.com/a/35744769
	type Alias Metadata
	return json.Marshal(&struct {
		*Alias
		FileDate string `json:",omitempty"`
	}{
		Alias:    (*Alias)(d),
		FileDate: d.FileDate.Format("02-Jan-2006 15:04:05"),
	})
}

type FileInfo struct { // for download, not authenticated: only safe info
	Description     string
	DaysUntilExpiry int
	ViewerCanDelete bool
	DownloadCount   int
}

type ErrorMessage struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

type SuccessMessage struct {
}

func readConfig(name string) Config {
	file, err := os.Open(name)
	if err != nil {
		log.Fatal("Error reading config: ", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Error reading config: ", err)
	}
	return config
}

func validateConfig(config Config) {
	if len(config.Listen) == 0 || len(config.StoragePath) == 0 || config.MaxFileSize <= 0 || config.MaxStorageSize <= 0 {
		log.Fatal("server.conf error")
	}
}

// serve files from go-bindata, print but ignore errors.
func serveFileAsset(w http.ResponseWriter, pathBelowClient string) {
	fmt.Println("serveFileAsset: " + pathBelowClient)
	mimeType := mime.TypeByExtension(filepath.Ext(pathBelowClient))
	b, err := Asset("client/" + pathBelowClient)
	if err != nil {
		log.Println(err)
	} else {
		w.Header().Set("Content-Type", mimeType)
		w.Write(b)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Println("index: url=" + r.URL.Path)
	if r.URL.Path == "/config.js" {
		http.ServeFile(w, r, "./config.js")
	} else {
		serveFileAsset(w, r.URL.Path[1:])
	}
}

func indexauth(w http.ResponseWriter, ar *auth.AuthenticatedRequest) {
	r := ar.Request
	fmt.Println("indexauth: url=" + r.URL.Path)
	if r.URL.Path == "/" {
		serveFileAsset(w, "upload.html")
	}
}

func download(w http.ResponseWriter, r *http.Request) {
	fmt.Println("download: url=" + r.URL.Path)
	serveFileAsset(w, "download.html")
}

func savaMetadata(identPath string, md Metadata) error {
	metaPath := identPath + ".json"
	// should erase unneeded fields from md, can't be done currently.
	metaContent, err := json.Marshal(md)
	if err != nil {
		return err
	}
	err = os.WriteFile(metaPath, metaContent, 0644)
	if err != nil {
		return err
	}
	return nil
}

type AdminInfo struct {
	Totalfilecount int
	Totalsize      int
}

type AdminFileList struct {
	FileList   []Metadata
	TotalSize  int64
	TotalFiles int
}

func admin(w http.ResponseWriter, ar *auth.AuthenticatedRequest) {
	r := ar.Request
	fmt.Println("admin: url=" + r.URL.Path + " query=" + r.URL.RawQuery)
	if r.URL.Path == "/admin/" {
		serveFileAsset(w, "admin.html")
		return
	} else if r.URL.Path == "/admin/get_files" {
		files, err := os.ReadDir(config.StoragePath) // TODO should be sorted...
		if err != nil {
			log.Println(err)
			return
		}
		afl := &AdminFileList{FileList: []Metadata{}, TotalSize: 0, TotalFiles: 0}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") {
				identPath := filepath.Join(config.StoragePath, strings.TrimSuffix(file.Name(), ".json"))
				md, err := loadMetadata(identPath)
				if err != nil {
					log.Println(err)
					return
				}
				afl.FileList = append(afl.FileList, md)
				afl.TotalFiles++
				afl.TotalSize += md.FileSize
			}
		}
		msg, _ := json.Marshal(afl)
		w.Write(msg)
		return
	} else if r.URL.Path == "/admin/delete_file" { // delete_file?filename
		ident := r.URL.RawQuery
		if ident != "" {
			identPath := path.Join(config.StoragePath, ident)
			deletefile(identPath, false)
		}
	}
}

func upload(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	fmt.Println("upload: url=" + r.Request.URL.Path)

	// check TotalSize
	start := time.Now()
	totalSize := int64(0)
	files, err := os.ReadDir(config.StoragePath)
	if err != nil {
		log.Println("Error opening I:", err, " path: ", config.StoragePath)
		return
	}
	for _, file := range files {
		inf, _ := file.Info()
		totalSize += inf.Size()
	}
	log.Printf("folder size took %s", time.Since(start))
	if totalSize > config.MaxStorageSize {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Storage size " + fmt.Sprintf("%d", config.MaxStorageSize) + " exceeded", Code: 1})
		w.Write(msg)
		return
	}

	if r.ContentLength > config.MaxFileSize {
		msg, _ := json.Marshal(&ErrorMessage{Error: "File size too large", Code: 1})
		w.Write(msg)
		return
	}

	r.ParseMultipartForm(50000000)
	file, _, err := r.FormFile("file")

	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: err.Error(), Code: 5})
		w.Write(msg)
		return
	}

	defer file.Close()

	ident := r.FormValue("ident")
	if len(ident) != 22 {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Ident filename length is incorrect", Code: 3})
		w.Write(msg)
		return
	}

	identPath := path.Join(config.StoragePath, path.Base(ident))
	if _, err := os.Stat(identPath); err == nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Ident is already taken.", Code: 4})
		w.Write(msg)
		return
	}

	// metadata unencrypted
	description := r.FormValue("description")
	expirydays, err := strconv.Atoi(r.FormValue("expirydays"))
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Expirydays not a number.", Code: 20})
		w.Write(msg)
		return
	}
	viewercandelete, err := strconv.ParseBool(r.FormValue("viewercandelete"))
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Viewercandelete not a bool.", Code: 21})
		w.Write(msg)
		return
	}
	savaMetadata(identPath, Metadata{Description: description, Expirydays: expirydays, Viewercandelete: viewercandelete})

	out, err := os.Create(identPath)
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: err.Error(), Code: 6})
		w.Write(msg)
		return
	}

	defer out.Close()

	out.Write([]byte{'U', 'P', '1', 0})
	_, err = io.Copy(out, file)
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: err.Error(), Code: 7})
		w.Write(msg)
		return
	}

	result, err := json.Marshal(&SuccessMessage{})
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: err.Error(), Code: 8})
		w.Write(msg)
	}
	w.Write(result)
}

func deletefile(identPath string, ignoreerrors bool) {
	if err := os.Remove(identPath); err != nil {
		if !ignoreerrors {
			log.Println(err)
		}
	}
	if err := os.Remove(identPath + ".json"); err != nil {
		if !ignoreerrors {
			log.Println(err)
		}
	}
}

func delfile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("delete: url=" + r.URL.Path)
	ident := r.FormValue("ident")

	if len(ident) != 22 {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Ident filename length is incorrect", Code: 3})
		w.Write(msg)
		return
	}

	identPath := path.Join(config.StoragePath, ident)
	if _, err := os.Stat(identPath); os.IsNotExist(err) {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Ident does not exist.", Code: 9})
		w.Write(msg)
		return
	}

	md, _ := loadMetadata(identPath)
	if !md.Viewercandelete {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Viewer can't delete.", Code: 91})
		w.Write(msg)
		return
	}

	deletefile(identPath, false)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func loadMetadata(identPath string) (Metadata, error) {
	file, err := os.ReadFile(identPath + ".json")
	md := Metadata{}
	if err != nil {
		return md, err
	}
	err = json.Unmarshal([]byte(file), &md)
	if ifile, err := os.Stat(identPath); err == nil {
		md.FileDate = ifile.ModTime()
		md.FileSize = ifile.Size()
		md.FileName = ifile.Name()
	}
	if md.Expirydays > 0 {
		md.DaysUntilExpiry = md.Expirydays - int(time.Since(md.FileDate).Hours()/24)
		if md.DaysUntilExpiry < 0 {
			md.DaysUntilExpiry = 0
		}
	} else {
		md.DaysUntilExpiry = -1
	}
	return md, err
}

func indexi(w http.ResponseWriter, r *http.Request) {
	fmt.Println("indexi: " + r.URL.Path)
	identPath := filepath.Join(config.StoragePath, r.URL.Path[2:])
	md, err := loadMetadata(identPath)
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Error loading metadata", Code: 30})
		w.Write(msg)
		return
	}
	md.Downloadcount++
	savaMetadata(identPath, md)
	fi := FileInfo{Description: md.Description, DaysUntilExpiry: md.DaysUntilExpiry, ViewerCanDelete: md.Viewercandelete, DownloadCount: md.Downloadcount}
	jsonstring, err := json.Marshal(fi)
	if err != nil {
		log.Println(err)
	}
	w.Header().Add("Fileinfo", string(jsonstring))

	http.ServeFile(w, r, identPath)
}

func expire() {
	time.Sleep(10 * time.Second)
	for {
		fmt.Println("expire: reading directory content...")
		files, err := os.ReadDir(config.StoragePath)
		if err != nil {
			log.Println(err)
		}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") {
				identPath := filepath.Join(config.StoragePath, strings.TrimSuffix(file.Name(), ".json"))
				md, err := loadMetadata(identPath)
				if err != nil {
					log.Println("expire: error loading metadata for " + identPath)
				} else {
					if md.DaysUntilExpiry == 0 {
						fmt.Println("expire: delete " + identPath)
						deletefile(identPath, true)
					}
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
		fmt.Println("expire: finished, sleeping...")
		time.Sleep(24 * time.Hour)
	}
}

func main() {
	configName := flag.String("config", "server.conf", "Configuration file")
	flag.Parse()

	config = readConfig(*configName)
	validateConfig(config)

	// http basic auth
	authenticator := auth.NewBasicAuthenticator("up1234me", auth.HtpasswdFileProvider("server.htpasswd"))

	http.HandleFunc("/i/", indexi)                     // serve encrypted files and unencrypted metadata
	http.HandleFunc("/up", authenticator.Wrap(upload)) // upload receiver
	http.HandleFunc("/del", delfile)
	http.HandleFunc("/d/", download) // download.html
	http.HandleFunc("/admin/", authenticator.Wrap(admin))
	http.HandleFunc("/", authenticator.Wrap(indexauth)) // this serves upload.html
	http.HandleFunc("/js/", index)
	http.HandleFunc("/deps/", index)
	http.HandleFunc("/favicon/", index)
	http.HandleFunc("/css/", index)
	http.HandleFunc("/config.js", index)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		log.Printf("Starting HTTP server on %s\n", config.Listen)
		log.Fatal("Error: ", http.ListenAndServe(config.Listen, nil))
	}()

	go expire() // run in parallel

	wg.Wait()
}
