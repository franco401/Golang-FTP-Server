# GolangFTPServer
An FTP client/server project written in Golang using the built-in net library. Contains both the server and client. Small sample files are added in a "files" folder so clients can download them. Clients are also able to upload files to the server as well. Use config.json to setup the ip address and port the server can host at, as well as the file buffer limit.

## Installation Guide

1. Download the source code

2. Build the client and server

```
go build client.go
```

```
go build server.go
```

## Usage
1. Use config.json to configure the ip address and port for the server to host on (currently defaults to "127.0.0.1" and "8080" respectively). This file also contains the file buffer limit for the client and server in bytes (currently defaults to 5242880 bytes or 5 MB)
2. Open the server and you will see the ip address and port number it's hosting on
3. Open the client, which will automatically connect to the server based on the configured ip address and port in config.json
4. Upon successfully connecting, clients can enter commands to do things such as view files stored in the "files" folder, download a file from the "files" folder, upload a file to the "files" folder, or exiting the server
5. If any kind of error occurs, the client is notified about it and exits the server

# Client commands
* vf = view files (from server)
* uf = upload files (to server)
* e = exit (close connection)

# File sizes for reference
* 1024 = 1 KB
* 1048576 = 1 MB
* 5242880 = 5 MB (default setting)
* 1073741824 = 1 GB
