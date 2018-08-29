package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	remoteAddr string
	bufSize    int64
)

func init() {
	flag.StringVar(&remoteAddr, "a", "", "Address of remote client")
	flag.Int64Var(&bufSize, "b", 1024, "Buffer size (default 1024)")
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
	start := time.Now()
	buf := make([]byte, bufSize)

	// Read file name, then file size, then file contents
	c.Read(buf)
	name := strings.Trim(string(buf), ":")

	c.Read(buf)
	size, _ := strconv.ParseInt(strings.Trim(string(buf), ":"), 10, 64)

	log.Printf("[info] getting %s (%d Mbytes) from %s\n", name, byteToMbyte(size), c.RemoteAddr().String())
	newFile, e := os.Create(name)
	if e != nil {
		log.Fatalf("[error] failed to create new file %s: %s\n", name, e.Error())
	}
	defer newFile.Close()

	// Start getting those bytes
	var received int64
	for {
		if (size - received) < bufSize {
			io.CopyN(newFile, c, (size - received))
			c.Read(make([]byte, (received+bufSize)-size))
			break
		}
		io.CopyN(newFile, c, bufSize)
		received += bufSize
	}
	tookSeconds := time.Now().Sub(start).Seconds()
	log.Printf("[info] got %s! took %.1fs\n", name, tookSeconds)
}

func byteToMbyte(b int64) int64 {
	return b / 1024 / 1024
}
