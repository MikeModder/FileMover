package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"time"
)

var (
	listenPort int
	bufferSize int

	file      string
	fileBytes []byte
	fileSize  int
)

func init() {
	flag.IntVar(&listenPort, "p", 1666, "TCP port (default 1666)")
	flag.IntVar(&bufferSize, "b", 1024, "Size of buffer (default 1024)")
	flag.StringVar(&file, "f", "", "File to send")
	flag.Parse()
	var e error

	if file == "" {
		log.Fatal("You must specofy a file!")
	}
	fileBytes, e = ioutil.ReadFile(file)
	if e != nil {
		log.Fatalf("[error] failed to read file '%s': %s\n", file, e.Error())
	}
	fileSize = len(fileBytes)
	log.Printf("[info] loaded %d Mbytes (%d bytes) from %s\n", byteToMbyte(int64(fileSize)), fileSize, file)
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
	// Send file metadata:
	//	Name
	//	Length
	start := time.Now()
	paddedName := pad(file)
	paddedLen := pad(strconv.Itoa(fileSize))

	// Write them
	_, e := c.Write([]byte(paddedName))
	if e != nil {
		log.Fatalf("[error] failed to write name to %s: %s\n", c.RemoteAddr().String(), e.Error())
	}
	_, e = c.Write([]byte(paddedLen))
	if e != nil {
		log.Fatalf("[error] failed to write len to %s: %s\n", c.RemoteAddr().String(), e.Error())
	}

	// Start writing file bytes
	buf := make([]byte, bufferSize)
	fileReader := bytes.NewReader(fileBytes)
	log.Printf("[info] starting to write file to %s\n", c.RemoteAddr().String())

	for {
		_, e := fileReader.Read(buf)
		if e == io.EOF {
			log.Println("[info] reached EOF")
			break
		}
		c.Write(buf)
	}
	tookSeconds := time.Now().Sub(start).Seconds()
	log.Printf("[info] file sent to %s, took %.1fs for %d Mbyte\n", c.RemoteAddr().String(), tookSeconds, byteToMbyte(int64(fileSize)))
	return
}

func pad(in string) string {
	for {
		l := len(in)
		if l < bufferSize {
			in = in + ":"
			continue
		}
		break
	}
	return in
}

func byteToMbyte(b int64) int64 {
	return b / 1024 / 1024
}
