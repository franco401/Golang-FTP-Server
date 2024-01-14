# GolangFTPServer
A simple FTP server project written in Golang using the built-in net library. Contains both the server and client. Small sample files are added in a "files" folder so clients can download them. Clients are also able to upload files to the server as well. Use config.json to setup the ip address and port the server can host at.

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
1. Use config.json to configure the ip address and port for the server to host on (ip address and port currently default to "127.0.0.1" and "8080")
2. Open the server
3. Open the client and you will enter the ip address and then the port so you can connect to the server
4. Upon successfully connecting, clients can enter commands to do things such as view files stored in the "files" folder, download a file from the "files" folder, upload a file to the "files" folder, or exiting the program
5. If a file is too large to download from or upload to the server respectively, the server notifies the client about it

# Client commands
* vf = view files (from server)
* uf = upload files (to server)
* e = exit (close connection)
