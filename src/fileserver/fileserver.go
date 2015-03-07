package main

import (
	"fmt"
	"net/http"
	"io"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	_ "strconv"
	"time"
	"mime"
)

var mux map[string]func(http.ResponseWriter, *http.Request)

type MyHandler struct {}
type home struct {
	Title string
	Addr  string
}

const (
	Template_Dir = "./view/"
	Upload_Dir   = "./"
)

func main() {
	server := &http.Server{
		Addr:    		":9090",
		Handler:    	&MyHandler{},
		ReadTimeout: 	10 * time.Second,
		WriteTimeout: 	10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	mux = make(map[string]func(http.ResponseWriter, *http.Request))
	mux["/"] = index
	mux["/upload"] = upload
	mux["/file"] = StaticServer
	server.ListenAndServe()
	fmt.Println(server.Addr)
}

func (*MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if h, ok := mux[r.URL.String()]; ok {
		h(w, r)
		return
	}
	if ok, _ := regexp.MatchString("/css/", r.URL.String()); ok {
		http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))).ServeHTTP(w, r)
	} else if ok, _ := regexp.MatchString("/js/", r.URL.String()); ok{
		http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))).ServeHTTP(w, r)
	}else {
		http.StripPrefix("/", http.FileServer(http.Dir("./upload/"))).ServeHTTP(w, r)
		setHeader(w)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles(Template_Dir + "upload.html")
		t.Execute(w, "上传文件")
	}else {
		r.ParseMultipartForm(32 << 20)
		file, handle, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Fprintf(w, "%v", "上传错误")
			return
		}
		fileext := filepath.Ext(handle.Filename)
		if check(fileext) == false {
			fmt.Fprintf(w, "%v", "不允许的上传类型")
			return
		}
		/*filename := strconv.FormatInt(time.Now().Unix(), 10) + fileext*/
		dir := r.FormValue("uploadpath")
		var uploadDir string
		if dir == "upload" {
			uploadDir = Upload_Dir + dir + "/"
		}else{
			uploadDir = Upload_Dir + "upload/" + dir + "/"
		}
		f, _ := os.OpenFile(uploadDir + handle.Filename, os.O_CREATE|os.O_WRONLY, 0660)
		defer f.Close()
		_, err = io.Copy(f, file)
		if err != nil {
			fmt.Fprintf(w, "%v", "上传失败")
			return
		}
		filedir, _ := filepath.Abs(uploadDir + handle.Filename)
		fmt.Fprintf(w, "%v", handle.Filename +"上传完成，服务器地址:"+filedir)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	title := home{Title : "首页"}
	t, _ := template.ParseFiles(Template_Dir + "index.html")
	t.Execute(w, title)
}

func StaticServer(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/file", http.FileServer(http.Dir("./upload"))).ServeHTTP(w, r)
	setHeader(w)
}

func check(name string) bool {
	ext := []string{".js"}
	for _, v := range ext {
		if v == name {
			return false
		}
	}
	return true
}

func setHeader(w http.ResponseWriter){
	fmt.Fprintln(w, "<title>文件浏览</title>")
	fmt.Fprintln(w, "<meta name='viewport' content='width=240,height=320,user-scalable=yes,initial-scale=2.5,maximum-scale=5.0,minimum-scale=1.0'>")
	/*设置mime,使android原生浏览器能够识别apk文件*/
	mime.AddExtensionType(".apk", "application/vnd.android.package-archive")
	/*w.Header().Set("Content-Type", "text/html;charset=utf-8")*/
	/*fmt.Fprintln(w, "<script type='text/javascript' src='http://127.0.0.1:9090/js/server.js' charset='utf-8'></script>")*/
}
