# FileMover Server
### Usage
`./Server -f SomeFile.zip` (uses 1024B chunks, and port `1666` by default)
### Advanced usage
`./Server -f <file name> -b <buffer size> -p <port to listen on>` (buffer size must be the same for both the client and server)
### What it does
It loads the file (into memory, so be careful with large ones!) then waits for a socket connection.

When it gets a connection it sends the file's name, then the file's size and then begins transferring the file (in 1024B chunks)