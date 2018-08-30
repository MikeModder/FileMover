# FileMover Client
### Usage
`./Client -a 192.168.0.14:1666`
### Advanced usage
`./Client -a <address> -b <buffer size>` (buffer size must be the same for both the client and server)
### What it does
It connects to the given address/port, and then waits for the file name and size. It creates and then writes the file. The file is recived in 1024B chunks by default.
### Building
To build this you need a working Go environment, [`github.com/dustin/go-humanize`](htt[github.com/dustin/go-humanize), and [`github.com/schollz/progressbar`](https://github.com/schollz/progressbar)