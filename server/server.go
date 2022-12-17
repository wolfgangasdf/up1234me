package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
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
	DeleteKey   string `json:"delete_key"`
	MaxFileSize int64  `json:"maximum_file_size"`

	Path struct {
		I      string `json:"i"`
		Client string `json:"client"`
	} `json:"path"`

	Http struct {
		Enabled bool   `json:"enabled"`
		Listen  string `json:"listen"`
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
	Delkey string `json:"delkey"`
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
	if !config.Http.Enabled {
		log.Fatal("Http must be enabled!")
	}
	if len(config.ApiKey) == 0 {
		log.Fatal("A static key must be defined in the configuration!")
	}
	if len(config.DeleteKey) == 0 {
		log.Fatal("A static delete key must be defined in the configuration!")
	}
	if len(config.Path.I) == 0 {
		config.Path.I = "../i"
	}
	if len(config.Path.Client) == 0 {
		config.Path.Client = "../client"
	}
}

func makeDelkey(ident string) string {
	key := []byte(config.DeleteKey)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(ident))
	return hex.EncodeToString(h.Sum(nil))
}

// func index(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
// 	fmt.Println("index: " + r.Username + " url=" + r.Request.URL.Path)
// 	// http.ServeFile(w, &r.Request, filepath.Join(config.Path.Client, "index.html"))
// 	if r.URL.Path == "/" {
// 		http.ServeFile(w, &r.Request, filepath.Join(config.Path.Client, "index.html"))
// 	} else {
// 		http.ServeFile(w, &r.Request, filepath.Join(config.Path.Client, r.URL.Path[1:]))
// 	}
// }

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
	metaPath := identPath + ".json"
	metaContent, err := json.Marshal(Metadata{Description: description, Expirydays: expirydays, Viewercandelete: viewercandelete})
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Metadata marshalling error:" + err.Error(), Code: 22})
		w.Write(msg)
		return
	}
	fmt.Println("writing metadata file... path: " + metaPath)
	err = ioutil.WriteFile(metaPath, metaContent, 0644)
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Error writing metadata file:" + err.Error(), Code: 23})
		w.Write(msg)
		return
	}

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

	delkey := makeDelkey(ident)

	result, err := json.Marshal(&SuccessMessage{Delkey: delkey})
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: err.Error(), Code: 8})
		w.Write(msg)
	}
	w.Write(result)
}

func delfile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("delete: url=" + r.URL.Path)
	ident := r.FormValue("ident")
	delkey := r.FormValue("delkey")

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

	if delkey != makeDelkey(ident) {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Incorrect delete key", Code: 10})
		w.Write(msg)
		return
	}

	os.Remove(identPath)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func indexi(w http.ResponseWriter, r *http.Request) {
	fmt.Println("indexi: " + r.URL.Path)
	identPath := filepath.Join(config.Path.I, r.URL.Path[2:])
	// read metadata
	file, err := ioutil.ReadFile(identPath + ".json")
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Error reading metadata file", Code: 30})
		w.Write(msg)
		return
	}
	md := Metadata{}
	err = json.Unmarshal([]byte(file), &md)
	if err != nil {
		msg, _ := json.Marshal(&ErrorMessage{Error: "Error decoding metadata", Code: 31})
		w.Write(msg)
		return
	}
	fi := FileInfo{Description: md.Description, DaysUntiExpiry: -1, ViewerCanDelete: md.Viewercandelete, DownloadCount: md.Downloadcount}
	if ifile, err := os.Stat(identPath); err != nil {
		fi.DaysUntiExpiry = int(time.Since(ifile.ModTime().AddDate(0, 0, md.Expirydays)).Hours() / 24)
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

	http.HandleFunc("/i/", indexi) // serve encrypted files and unencrypted metadata
	http.HandleFunc("/up", authenticator.Wrap(upload))
	http.HandleFunc("/del", delfile)
	http.HandleFunc("/d/", download) // download screen
	http.HandleFunc("/", index)      // serve all other files

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if config.Http.Enabled {
			log.Printf("Starting HTTP server on %s\n", config.Http.Listen)
			log.Println(http.ListenAndServe(config.Http.Listen, nil))
		}
	}()

	wg.Wait()
}
