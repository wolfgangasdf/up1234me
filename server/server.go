package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
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
	Listen      string `json:"listen"`
	MaxFileSize int64  `json:"maximum_file_size"`

	Path struct {
		I      string `json:"i"`
		Client string `json:"client"`
	} `json:"path"`

	Http struct {
		Listen string `json:"listen"`
	} `json:"http"`
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
}

type FileInfo struct {
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
	file, _ := os.Open(name)
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Config{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error reading config: ", err)
	}
	return config
}

func validateConfig(config Config) {
	if len(config.Path.I) == 0 {
		config.Path.I = "../i"
	}
	if len(config.Path.Client) == 0 {
		config.Path.Client = "../client"
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Println("index: url=" + r.URL.Path)
	http.ServeFile(w, r, filepath.Join(config.Path.Client, r.URL.Path[1:]))
}

func indexauth(w http.ResponseWriter, ar *auth.AuthenticatedRequest) {
	r := ar.Request
	fmt.Println("indexauth: url=" + r.URL.Path)
	if r.URL.Path == "/" {
		http.ServeFile(w, &r, filepath.Join(config.Path.Client, "upload.html"))
	}
}

func download(w http.ResponseWriter, r *http.Request) {
	fmt.Println("download: url=" + r.URL.Path)
	http.ServeFile(w, r, filepath.Join(config.Path.Client, "download.html"))
}

func savaMetadata(identPath string, md Metadata) error {
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
	FileList []Metadata
}

func admin(w http.ResponseWriter, ar *auth.AuthenticatedRequest) {
	r := ar.Request
	fmt.Println("admin: url=" + r.URL.Path)
	if r.URL.Path == "/admin/" {
		http.ServeFile(w, &r, filepath.Join(config.Path.Client, "admin.html"))
		return
	} else if r.URL.Path == "/admin/get_info" {
		msg, _ := json.Marshal(&AdminInfo{Totalfilecount: 123, Totalsize: 321}) // TODO
		w.Write(msg)
		return
	} else if r.URL.Path == "/admin/get_files" {
		fmt.Println("admin: get_file: " + r.FormValue("startindex"))

		files, err := os.ReadDir(config.Path.I)
		if err != nil {
			log.Fatal(err)
		}
		afl := &AdminFileList{FileList: []Metadata{}}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") {
				identPath := filepath.Join(config.Path.I, strings.TrimSuffix(file.Name(), ".json"))
				md, err := loadMetadata(identPath)
				if err != nil {
					log.Fatal(err)
				}
				afl.FileList = append(afl.FileList, md)
			}
		}
		msg, _ := json.Marshal(afl)
		w.Write(msg)
		return
	} else if r.URL.Path == "/admin/delete_all_before" {
		// TODO
	} else if r.URL.Path == "/admin/delete_file" {
		// TODO
	}
}

func upload(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	fmt.Println("upload: url=" + r.Request.URL.Path)
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

	identPath := path.Join(config.Path.I, path.Base(ident))
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
			panic(err)
		}
	}
	if err := os.Remove(identPath + ".json"); err != nil {
		if !ignoreerrors {
			panic(err)
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

	identPath := path.Join(config.Path.I, ident)
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
	identPath := filepath.Join(config.Path.I, r.URL.Path[2:])
	md, err := loadMetadata(identPath)
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Error loading metadata", Code: 30})
		w.Write(msg)
		return // TODO test this, does client receive error?
	}
	md.Downloadcount++
	savaMetadata(identPath, md)
	fi := FileInfo{Description: md.Description, DaysUntilExpiry: md.DaysUntilExpiry, ViewerCanDelete: md.Viewercandelete, DownloadCount: md.Downloadcount}
	jsonstring, err := json.Marshal(fi)
	if err != nil {
		panic(err)
	}
	w.Header().Add("Fileinfo", string(jsonstring))

	http.ServeFile(w, r, identPath)
}

func expire() {
	time.Sleep(10 * time.Second)
	for {
		fmt.Println("expire: reading directory content...")
		files, err := os.ReadDir(config.Path.I)
		if err != nil {
			log.Fatal(err)
		}
		var iii = 0
		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".json") {
				iii++
				identPath := filepath.Join(config.Path.I, file.Name())
				if iii%int(math.Ceil(float64(len(files))/10.0)) == 0 {
					fmt.Println("expire: checking " + identPath)
				}
				md, err := loadMetadata(identPath)
				if err != nil {
					fmt.Println("expire: error loading metadata for " + identPath)
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
	wg.Add(2)

	go func() {
		defer wg.Done()
		log.Printf("Starting HTTP server on %s\n", config.Http.Listen)
		log.Println(http.ListenAndServe(config.Http.Listen, nil))
	}()

	go expire() // run in goroutine

	wg.Wait()
}
