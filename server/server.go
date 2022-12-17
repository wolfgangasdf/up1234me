package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	auth "github.com/abbot/go-http-auth"
)

type Config struct {
	Listen      string `json:"listen"`
	ApiKey      string `json:"api_key"`
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
	Viewercandelete bool
	Downloadcount   int
}

type FileInfo struct {
	Description     string
	DaysUntiExpiry  int
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
	decoder := json.NewDecoder(file)
	config := Config{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error reading config: ", err)
	}
	return config
}

func validateConfig(config Config) {
	if len(config.ApiKey) == 0 {
		log.Fatal("A static key must be defined in the configuration!")
	}
	if len(config.Path.I) == 0 {
		config.Path.I = "../i"
	}
	if len(config.Path.Client) == 0 {
		config.Path.Client = "../client"
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Println("index: url=" + r.URL.Path)
	if r.URL.Path == "/" {
		http.ServeFile(w, r, filepath.Join(config.Path.Client, "upload.html"))
	} else {
		http.ServeFile(w, r, filepath.Join(config.Path.Client, r.URL.Path[1:]))
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

	apikey := r.FormValue("api_key")
	if apikey != config.ApiKey {
		msg, _ := json.Marshal(&ErrorMessage{Error: "API key doesn't match", Code: 2})
		w.Write(msg)
		return
	}

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
	fmt.Println("description: " + description + " expirydays=" + strconv.Itoa(expirydays) + " viewercandelete=" + strconv.FormatBool(viewercandelete))
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

	if err := os.Remove(identPath); err != nil {
		panic(err)
	}
	if err := os.Remove(identPath + ".json"); err != nil {
		panic(err)
	}
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func loadMetadata(identPath string) (Metadata, error) {
	file, err := os.ReadFile(identPath + ".json")
	md := Metadata{}
	if err != nil {
		return md, err
	}
	err = json.Unmarshal([]byte(file), &md)
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
	fi := FileInfo{Description: md.Description, DaysUntiExpiry: -1, ViewerCanDelete: md.Viewercandelete, DownloadCount: md.Downloadcount}
	if ifile, err := os.Stat(identPath); err == nil {
		fi.DaysUntiExpiry = md.Expirydays - int(time.Since(ifile.ModTime()).Hours()/24)
	}
	jsonstring, err := json.Marshal(fi)
	if err != nil {
		panic(err)
	}
	w.Header().Add("Fileinfo", string(jsonstring))

	http.ServeFile(w, r, identPath)
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
	http.HandleFunc("/", index)      // serve all other files

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		log.Printf("Starting HTTP server on %s\n", config.Http.Listen)
		log.Println(http.ListenAndServe(config.Http.Listen, nil))
	}()

	wg.Wait()
}
