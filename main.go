package main

import (
	"./core"
	"log"
	"net/http"
	"io"
	"os"
	"crypto/md5"
	"path/filepath"
	"time"
	"strings"
	"strconv"
	"encoding/hex"
	"runtime"
	"net/url"
	"syscall"
	"encoding/json"
	"io/ioutil"
	"bytes"
	"image/jpeg"
	_"image/png"
)

const (
	UPLOAD_DIR = "/tmp/image"
)

type ImageProfile struct {
	Path          string
	Width, Height int
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		var width, height int
		if queryForm, err := url.ParseQuery(r.URL.RawQuery); err == nil {
			if len(queryForm["w"]) > 0 {
				width, _ = strconv.Atoi(queryForm["w"][0])
			} else {
				width = 250
			}
			if len(queryForm["h"]) > 0 {
				height, _ = strconv.Atoi(queryForm["h"][0])
			} else {
				height = 300
			}
		}

		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		//文件目录 按日期聚合
		md5h := md5.New()
		io.Copy(md5h, file)
		name := hex.EncodeToString(md5h.Sum([]byte("")))
		now := time.Now()
		ext := strings.ToLower(filepath.Ext(handler.Filename))
		path := filepath.Join(strconv.Itoa(now.Year()), strconv.Itoa(int(now.Month())), strconv.Itoa(now.Day()))

		os.MkdirAll(filepath.Join(UPLOAD_DIR, path), os.ModePerm)

		originFile, err := os.Create(filepath.Join(UPLOAD_DIR, path, name + "_origin" + ext))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer originFile.Close()
		file.Seek(0, 0)
		io.Copy(originFile, file)

		file.Seek(0, 0)
		thumb, err := core.Crop(file, core.Config{Width: width, Height: height})

		result := filepath.Join(path, name + "_" + strconv.Itoa(width) + "x" + strconv.Itoa(height) + ext)
		thumbPath := filepath.Join(UPLOAD_DIR, result)
		out, err := os.Create(thumbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer out.Close()

		buffer := new(bytes.Buffer)
		if err := jpeg.Encode(buffer, thumb, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = ioutil.WriteFile(thumbPath, buffer.Bytes(), 0666)

		profile, err := json.Marshal(ImageProfile{result, width, height})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Server", "GoCrop")
		w.WriteHeader(http.StatusOK)
		w.Write(profile)
	} else {
		http.Error(w, "Request method [" + r.Method + "] not supported", http.StatusMethodNotAllowed);
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	mask := syscall.Umask(0)
	defer syscall.Umask(mask)

	http.HandleFunc("/upload", upload)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}