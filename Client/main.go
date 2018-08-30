package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/schollz/progressbar"
)

var (
	remoteAddr  string
	initBufSize int64
	bufSize     int64
	files       []fileinfo
)

type config struct {
	BufferSize int64  `json:"buffersize"`
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
	flag.StringVar(&remoteAddr, "a", "", "Address of remote client")
	flag.Int64Var(&initBufSize, "b", 1024, "Starting buffer size (default 1024)")
	flag.Parse()
}

func main() {
	// Try and connect to the server
	c, e := net.Dial("tcp", remoteAddr)
	if e != nil {
		log.Fatalf("[error] failed to dail remote server: %s\n", remoteAddr)
	}
	defer c.Close()

	// Connection all good(?)
	confBuf := make([]byte, initBufSize)

	// Read config
	c.Read(confBuf)
	cfgBytes := []byte(strings.Trim(string(confBuf), "|"))
	var cfg config
	e = json.Unmarshal(cfgBytes, &cfg)
	if e != nil {
		log.Fatalf("[error] failed to unmarshal JSON: %s\n", e.Error())
	}

	log.Printf("Server MOTD: %s\n%d files, using %d chunks\n", cfg.MOTD, cfg.FileCount, cfg.BufferSize)
	bufSize = cfg.BufferSize

	log.Println("remote files:")
	buf := make([]byte, bufSize)
	for i := 0; i < cfg.FileCount; i++ {
		// Get i file info packets
		_, e := c.Read(buf)
		if e != nil {
			log.Fatalf("[error] failed to read: %s\n", e.Error())
		}

		f := fileinfo{}
		fiBytes := []byte(strings.Trim(string(buf), "|"))
		json.Unmarshal(fiBytes, &f)
		log.Printf("%s - %s - ID: %d", f.Name, humanize.Bytes(uint64(f.Size)), f.ID)
		files = append(files, f)
	}
	var id int

	// get user input for file id
	var input string
	for {
		log.Printf("Choose a file [%d-%d]: ", 0, len(files)-1)
		fmt.Scanln(&input)
		id, e = strconv.Atoi(input)
		if e != nil {
			log.Println("That wasn't a valid number, try again!")
			continue
		}
		if id > len(files)-1 || id < 0 {
			log.Println("We don't have a file with that ID, try again!")
			continue
		}
		break
	}
	log.Printf("[info] getting file with id %d\n", id)

	c.Write([]byte(pad(strconv.Itoa(id))))
	log.Printf("[info] getting %s (%s)\n", files[id].Name, humanize.Bytes(uint64(files[id].Size)))

	bar := progressbar.New(int(files[id].Size))

	newFile, e := os.Create(files[id].Name)
	if e != nil {
		log.Fatalf("[error] failed to create new file %s: %s\n", files[id].Name, e.Error())
	}
	defer newFile.Close()

	// Start getting those bytes
	var received int64
	for {
		if (files[id].Size - received) < bufSize {
			io.CopyN(newFile, c, (files[id].Size - received))
			c.Read(make([]byte, (received+bufSize)-files[id].Size))
			bar.Finish()
			break
		}
		io.CopyN(newFile, c, bufSize)
		bar.Add(int(bufSize))
		received += bufSize
	}
}

func pad(in string) string {
	for {
		l := len(in)
		if l < int(bufSize) {
			in = in + "|"
			continue
		}
		break
	}
	return in
}
