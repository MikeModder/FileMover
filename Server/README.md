# FileMover Server
### Usage
`./Server` (uses 1024B chunks, current directory, and port `1666` by default)
### Advanced usage
`./Server -f <folder path> -b <buffer size> -p <port to listen on>` (buffer size must be the same for both the client and server)
### Building
To build this project, you need a working Go environment, [`github.com/JoshuaDoes/govvv`](https://github.com/JoshuaDoes/govvv) and [`github.com/dustin/go-humanize`](htt[github.com/dustin/go-humanize). Once you have all that, simply run `govvv build`.
### What it does
When a client connects it sends a packet with Server MOTD, file count and buffer size. It then waits for the client to send an ID. The server then sends the file with that ID.