package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
)

var (
	listenPort int
	bufferSize int
	motd       string

	cfg config

	dir       string
	files     []fileinfo
	fileBytes []byte
	fileSize  int

	//GitCommit git commit
	GitCommit string
	//BuildDate build date
	BuildDate string
)

type config struct {
	BufferSize int    `json:"buffersize"`
	MOTD       string `json:"motd"`
	FileCount  int    `json:"filecount"`
}

type fileinfo struct {
	Name     string `json:"name"`
	FullPath string
	Size     int64 `json:"size"`
	ID       int   `json:"id"`
}

func init() {
	flag.IntVar(&listenPort, "p", 1666, "TCP port")
	flag.IntVar(&bufferSize, "b", 1024, "Size of buffer")
	flag.StringVar(&motd, "m", "FileMover server git-"+GitCommit+"-"+BuildDate, "Optional MOTD for server")
	flag.StringVar(&dir, "d", ".", "Directory with files to send")
	flag.Parse()
	var e error

	dirFiles, e := ioutil.ReadDir(dir)
	if e != nil {
		log.Fatalf("[error] failed to read dir '%s': %s\n", dir, e.Error())
	}

	if len(dirFiles) == 0 {
		log.Fatalf("[error] no files in directory %s\n", dir)
	}

	for i := 0; i < len(dirFiles); i++ {
		f := dirFiles[i]
		if f.IsDir() {
			continue
		}
		log.Printf("%s - %s", f.Name(), humanize.Bytes(uint64(f.Size())))
		files = append(files, fileinfo{
			Name:     f.Name(),
			FullPath: path.Join(dir, f.Name()),
			Size:     f.Size(),
			ID:       i,
		})
	}

	cfg = config{
		BufferSize: bufferSize,
		MOTD:       motd,
		FileCount:  len(files),
	}
	log.Printf("%s - %d buffer, %d files\n", cfg.MOTD, cfg.BufferSize, cfg.FileCount)
}

func main() {
	addr := &net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: listenPort,
	}
	server, e := net.ListenTCP("tcp", addr)
	if e != nil {
		log.Fatalf("[error] can't listen on port %d: %s\n", listenPort, e.Error())
	}
	defer server.Close()

	for {
		nc, e := server.Accept()
		if e != nil {
			log.Printf("[error] failed to accept connection from %s: %s\n", nc.RemoteAddr().String(), e.Error())
		}

		log.Printf("[info] remote %s connected\n", nc.RemoteAddr().String())
		go handleConn(nc)
	}
}

func handleConn(c net.Conn) {
	// First send config packet, then send file info
	// Marshall the config object
	confBytes, e := json.Marshal(cfg)
	if e != nil {
		log.Fatalf("[error] failed to marshall JSON: %s\n", e.Error())
	}
	paddedConfBytes := []byte(pad(string(confBytes)))

	// Write it to the socket
	_, e = c.Write(paddedConfBytes)
	if e != nil {
		log.Printf("[error] failed to write config to remote %s: %s\n", c.RemoteAddr().String(), e.Error())
		return
	}

	// Loop through files and send them to the client
	for i := 0; i < len(files); i++ {
		f := files[i]
		fb, _ := json.Marshal(f)
		paddedFB := []byte(pad(string(fb)))
		c.Write(paddedFB)
	}

	// Make the incoming buffer
	buf := make([]byte, bufferSize)

	// Wait for the client to send the ID of the file they want
	_, e = c.Read(buf)
	if e != nil {
		log.Printf("[error] failed to read id from remote %s: %s\n", c.RemoteAddr().String(), e.Error())
		return
	}

	wantedID, e := strconv.Atoi(strings.Trim(string(buf), "|"))
	if e != nil {
		log.Printf("[error] failed to parse id '%s' from remote %s: %s\n", strings.Trim(string(buf), "|"), c.RemoteAddr().String(), e.Error())
		return
	}

	if wantedID > len(files) || wantedID < 0 {
		log.Printf("[error] remote %s requested invalid id '%d'\n", c.RemoteAddr().String(), wantedID)
		c.Close()
		return
	}

	if !writeFile(files[wantedID].FullPath, c) {
		log.Printf("[error] failed to send %s to remote %s\n", files[wantedID].Name, c.RemoteAddr().String())
	}

	return
}

func writeFile(path string, to net.Conn) bool {
	buf := make([]byte, bufferSize)
	file, e := os.Open(path)
	defer file.Close()
	if e != nil {
		log.Printf("[error] failed to open %s: %s\n", path, e.Error())
		return false
	}
	log.Printf("[info] starting to write file to %s\n", to.RemoteAddr().String())

	for {
		_, e := file.Read(buf)
		if e == io.EOF {
			log.Println("[info] reached EOF")
			break
		}
		to.Write(buf)
	}
	return true
}

func pad(in string) string {
	for {
		l := len(in)
		if l < bufferSize {
			in = in + "|"
			continue
		}
		break
	}
	return in
}
