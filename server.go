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
	"sort"
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

type MetadataSaved struct {
	Description     string
	Expirydays      int
	Viewercandelete bool
	Downloadcount   int
}

type MetadataTemp struct {
	Saved           MetadataSaved
	FileDate        time.Time
	FileSize        int64
	FileName        string
	DaysUntilExpiry int
}

func (d *MetadataTemp) MarshalJSON() ([]byte, error) { // format date https://stackoverflow.com/a/35744769
	type Alias MetadataTemp
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
func serveStaticFile(w http.ResponseWriter, pathBelowClient string) {
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
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/upload", http.StatusMovedPermanently)
	} else if r.URL.Path == "/config.js" {
		http.ServeFile(w, r, "./config.js")
	} else {
		serveStaticFile(w, r.URL.Path[1:])
	}
}

func uploadhtml(w http.ResponseWriter, ar *auth.AuthenticatedRequest) {
	r := ar.Request
	fmt.Println("uploadhtml: url=" + r.URL.Path)
	if r.URL.Path == "/upload" {
		serveStaticFile(w, "upload.html")
	}
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
	md.Saved.Downloadcount++
	savaMetadata(identPath, md.Saved)
	fi := FileInfo{Description: md.Saved.Description, DaysUntilExpiry: md.DaysUntilExpiry, ViewerCanDelete: md.Saved.Viewercandelete}
	jsonstring, err := json.Marshal(fi)
	if err != nil {
		log.Println(err)
	}
	w.Header().Add("Fileinfo", string(jsonstring))

	http.ServeFile(w, r, identPath)
}

func download(w http.ResponseWriter, r *http.Request) {
	fmt.Println("download: url=" + r.URL.Path)
	serveStaticFile(w, "download.html")
}

func savaMetadata(identPath string, md MetadataSaved) error {
	metaPath := identPath + ".json"
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
	FileList   []MetadataTemp
	TotalSize  int64
	TotalFiles int
}

func admin(w http.ResponseWriter, ar *auth.AuthenticatedRequest) {
	r := ar.Request
	fmt.Println("admin: url=" + r.URL.Path + " query=" + r.URL.RawQuery)
	if r.URL.Path == "/admin/" {
		serveStaticFile(w, "admin.html")
		return
	} else if r.URL.Path == "/admin/get_files" {
		files, err := os.ReadDir(config.StoragePath)
		if err != nil {
			log.Println(err)
			return
		}
		afl := &AdminFileList{FileList: []MetadataTemp{}, TotalSize: 0, TotalFiles: 0}
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
		sort.Slice(afl.FileList, func(i, j int) bool {
			return afl.FileList[i].FileDate.After(afl.FileList[j].FileDate)
		})
		msg, _ := json.Marshal(afl)
		w.Write(msg)
		return
	} else if r.URL.Path == "/admin/delete_file" { // delete_file?filename
		ident := r.URL.RawQuery
		if ident != "" {
			identPath := path.Join(config.StoragePath, ident)
			deletefile(identPath, false)
		}
		result, _ := json.Marshal(&SuccessMessage{})
		w.Write(result)
	} else if r.URL.Path == "/admin/set_expiry" { // set_expiry?i=ident&days=123
		ident := r.URL.Query().Get("i")
		days, err := strconv.Atoi(r.URL.Query().Get("days"))
		if err != nil {
			log.Println("set_expiry: Error parsing days: ", err)
			return
		}
		log.Println("set_expiry: i=", ident, " days=", days)
		identPath := filepath.Join(config.StoragePath, ident)
		md, err := loadMetadata(identPath)
		if err != nil {
			log.Println("Error loading metadata: ", err)
			return
		}
		md.Saved.Expirydays = days
		savaMetadata(identPath, md.Saved)
	} else if r.URL.Path == "/admin/set_viewercandelete" { // set_viewercandelete?i=ident&b=[true|false]
		ident := r.URL.Query().Get("i")
		b, err := strconv.ParseBool(r.URL.Query().Get("b"))
		if err != nil {
			log.Println("set_viewercandelete: Error parsing bool: ", err)
			return
		}
		log.Println("set_viewercandelete: i=", ident, " b=", b)
		identPath := filepath.Join(config.StoragePath, ident)
		md, err := loadMetadata(identPath)
		if err != nil {
			log.Println("Error loading metadata: ", err)
			return
		}
		md.Saved.Viewercandelete = b
		savaMetadata(identPath, md.Saved)
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
		msg, _ := json.Marshal(&ErrorMessage{Error: "Storage size exceeded", Code: 1})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(msg)
		return
	}

	if r.ContentLength > config.MaxFileSize {
		msg, _ := json.Marshal(&ErrorMessage{Error: "File size too large", Code: 1})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(msg)
		return
	}

	r.ParseMultipartForm(config.MaxFileSize + 1024)
	file, _, err := r.FormFile("file")

	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: err.Error(), Code: 5})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(msg)
		return
	}

	defer file.Close()

	ident := r.FormValue("ident")
	if len(ident) != 22 {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Ident filename length is incorrect", Code: 3})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(msg)
		return
	}

	identPath := path.Join(config.StoragePath, path.Base(ident))
	if _, err := os.Stat(identPath); err == nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Ident is already taken.", Code: 4})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(msg)
		return
	}

	// metadata unencrypted
	description := r.FormValue("description")
	expirydays, err := strconv.Atoi(r.FormValue("expirydays"))
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Expirydays not a number.", Code: 20})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(msg)
		return
	}
	viewercandelete, err := strconv.ParseBool(r.FormValue("viewercandelete"))
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Viewercandelete not a bool.", Code: 21})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(msg)
		return
	}
	savaMetadata(identPath, MetadataSaved{Description: description, Expirydays: expirydays, Viewercandelete: viewercandelete})

	out, err := os.Create(identPath)
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: err.Error(), Code: 6})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(msg)
		return
	}

	defer out.Close()

	out.Write([]byte{'U', 'P', '1', 0})
	_, err = io.Copy(out, file)
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: err.Error(), Code: 7})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(msg)
		return
	}

	result, _ := json.Marshal(&SuccessMessage{})
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
	if !md.Saved.Viewercandelete {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Viewer can't delete.", Code: 91})
		w.Write(msg)
		return
	}

	deletefile(identPath, false)
}

func loadMetadata(identPath string) (MetadataTemp, error) {
	file, err := os.ReadFile(identPath + ".json")
	md := MetadataTemp{}
	if err != nil {
		return md, err
	}
	err = json.Unmarshal([]byte(file), &md.Saved)
	if ifile, err := os.Stat(identPath); err == nil {
		md.FileDate = ifile.ModTime()
		md.FileSize = ifile.Size()
		md.FileName = ifile.Name()
	}
	if md.Saved.Expirydays > 0 {
		md.DaysUntilExpiry = md.Saved.Expirydays - int(time.Since(md.FileDate).Hours()/24)
		if md.DaysUntilExpiry < 0 {
			md.DaysUntilExpiry = 0
		}
	} else {
		md.DaysUntilExpiry = -1
	}
	return md, err
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

	http.HandleFunc("/", index)                                // serve static files or redirect / to upload
	http.HandleFunc("/d/", download)                           // download.html
	http.HandleFunc("/i/", indexi)                             // serve encrypted files and unencrypted metadata
	http.HandleFunc("/del", delfile)                           // downloader delete
	http.HandleFunc("/upload", authenticator.Wrap(uploadhtml)) // this serves upload.html
	http.HandleFunc("/up", authenticator.Wrap(upload))         // upload receiver
	http.HandleFunc("/admin/", authenticator.Wrap(admin))      // admin

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
