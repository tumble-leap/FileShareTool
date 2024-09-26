package main

import (
	"encoding/json"
	"errors"
	"file-share-tool/webapp"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

type File struct {
	FileName    string `json:"file_name"`
	FileModtime string `json:"file_modtime"`
	IsDir       bool   `json:"is_dir"`
	FileSize    string `json:"file_size"`
	SubFileNum  int    `json:"sub_file_num"`
	SubDirNum   int    `json:"sub_dir_num"`
}

type Resp struct {
	Message    string `json:"message"`
	LocalIP    string `json:"local_ip"`
	TargetPath string `json:"path"`
	FileList   []File `json:"files"`
}

var IgnoreList = []string{
	".DS_Store",
	".Trash",
	".localized",
}

func GetInternalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", errors.New("internal IP fetch failed, detail:" + err.Error())
	}
	defer conn.Close()
	res := conn.LocalAddr().String()
	res = strings.Split(res, ":")[0]
	return res, nil
}

func humanReadableSize(size int64) string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
		TB = 1 << 40
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/TB)
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func CountDirsAndFiles(path string) (dirs, files int, err error) {
	fs, err := os.ReadDir(path)
	if err != nil {
		return 0, 0, err
	}
	for _, v := range fs {
		info, _ := v.Info()
		if slices.Contains[[]string, string](IgnoreList, info.Name()) {
			continue
		}
		if info.IsDir() {
			dirs++
		} else {
			files++
		}
	}
	return
}

var (
	// homeDir    string
	currentDir string
	localIp    string
	port       = 8000
)

func getFileListHandler(w http.ResponseWriter, r *http.Request) {
	targetDir := r.URL.Query().Get("path")
	if targetDir == "" || targetDir == "undefined" {
		targetDir = currentDir
	}

	var resp = Resp{
		TargetPath: targetDir,
	}
	resp.LocalIP = localIp + ":" + "8000"

	fs, err := os.ReadDir(targetDir)
	if err != nil {
		resp.Message = err.Error()
	} else {
		var dirs, files []File
		for _, v := range fs {
			info, _ := v.Info()
			if slices.Contains[[]string, string](IgnoreList, info.Name()) {
				continue
			}
			var f File
			f.FileName = info.Name()
			f.FileModtime = info.ModTime().Local().Format("2006/01/02 15:04:05")
			if info.IsDir() {
				dNum, fNum, _ := CountDirsAndFiles(filepath.Join(targetDir, info.Name()))
				f.SubDirNum = dNum
				f.SubFileNum = fNum
				f.IsDir = true
				dirs = append(dirs, f)
			} else {
				f.FileSize = humanReadableSize(info.Size())
				files = append(files, f)
			}
		}
		resp.FileList = append(resp.FileList, dirs...)
		resp.FileList = append(resp.FileList, files...)
	}

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	log.Println("request dir: ", targetDir)
	json.NewEncoder(w).Encode(&resp)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	targetFile := r.URL.Query().Get("fname")
	fmt.Println("request download file:", targetFile)
	if targetFile == "" {
		return
	}

	fileName := filepath.Base(targetFile)

	// 打开文件
	file, err := os.Open(targetFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error opening file: %s", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 设置响应头
	// w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	// w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	// w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")

	// 将文件内容写入响应体
	_, err = io.Copy(w, file)
	log.Println("request download file: ", targetFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error copying file to response: %s", err), http.StatusInternalServerError)
		return
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	case "linux": // Linux
		cmd = exec.Command("xdg-open", url)
	case "windows": // Windows
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		fmt.Println("Unsupported platform")
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("Error opening browser:", err)
	}
}

func main() {

	// get home dir
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	homeDir := currentUser.HomeDir

	currentDir, _ = os.Getwd()

	// command line tool
	var cmdlDir string
	flag.StringVar(&cmdlDir, "t", currentDir, "The file server root directory, the default is the current directory")
	flag.Parse()

	if cmdlDir != "" {
		if cmdlDir == "home" {
			currentDir = homeDir
		} else {
			currentDir = cmdlDir
		}
	}

	// get ip
	localIpAddr, err := GetInternalIP()
	if err != nil {
		log.Println(err)
		log.Println("并没有获取到本机的ip地址呢，请手动查询～")
	} else {
		log.Println("本机ip：" + localIpAddr)
		localIp = localIpAddr
	}

	// 设置路由和处理函数
	http.HandleFunc("/api/files", getFileListHandler)
	http.HandleFunc("/api/download", downloadHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 在这里处理React项目的静态文件
		fs, err := webapp.FS()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.FileServer(fs).ServeHTTP(w, r)
	})

	// 启动服务器并监听端口

	log.Printf("Server is running on http://localhost:%d\n", port)
	log.Printf("Please open http://localhost:%d in this PC\n", port)
	log.Printf("Or open http://%s:%d in other PC under the LAN\n", localIp, port)
	go openBrowser(fmt.Sprintf("http://localhost:%d", port))
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
