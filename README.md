# GolangFTPServer
An FTP client/server project written in Golang using the built-in net library. Contains both the server and client. Small sample files are added in a "files" folder so clients can download them (you can delete this folder if you like). The client program can upload files to the server program and download files from it. Use config.json to setup the ip address and port the server can host at, the max file buffer size limit and the server's file storage directory.

# What is the max file buffer size?
* This is the amount of bytes to download and receive for a file in chunks. This is to enable memory efficient file transferring as only chunks of file data is in memory and sent instead of the entire file at once. This setting currently defaults to 5242880 bytes or 5 MB but you can change it to whatever you like.

# What is the file storage directory?
* This is the folder where the client program can upload files to and where the server reads all file names and sizes to share with the client program when they enter the "vf" command. Currently defaults to "./files/" To use a different folder, in config.json, change "./files/" to "./[your_folder_name]/"

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
1. Use config.json to configure the ip address and port for the server to host on (currently defaults to "127.0.0.1" and "8080" respectively), the max file buffer size for the client and server in bytes, and the file storage directory.
2. Open the server and you will see the ip address and port number it's hosting on
3. Open the client, which will automatically connect to the server based on the configured ip address and port in config.json
4. Upon successfully connecting, clients can enter commands to do things such as view files stored in the "files" folder, download a file from the "files" folder, upload a file to the "files" folder, or exiting the server
5. If any kind of error occurs, the client is notified about it and exits the server. If the client enters a filename that can't be read/found clientside or serverside (for file uploading/downloading respectively), a 0 KB file will appear, just delete it as it is empty.

# Client commands
* vf = view files (from server)
* uf = upload files (to server)
* e = exit (close connection)

# File sizes for reference
* 1024 = 1 KB
* 1048576 = 1 MB
* 5242880 = 5 MB (default setting)
* 1073741824 = 1 GB
