package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

func GetFileSize(filename string) string {
	//read file
	file, err := os.Open(filename)
	//close file at the end of current file goroutine
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}

	f, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	//file size in bytes
	fs := f.Size()
	var filesize float64 = float64(fs)

	filesizes := []string{"KB", "MB", "GB"}
	var filesize_index int8

	/*
	* while the filesize in bytes
	* is bigger than 1024, continously
	* divide by 1024 and increment the
	* filesize_index to finally return
	* the appropriate file size for the
	* client, such as 1 KB for 1024 bytes
	 */
	for filesize >= 1024.0 {
		filesize_index++
		filesize /= 1024.0
	}
	return fmt.Sprintln(filesize, filesizes[filesize_index])
}

func GetFileData() string {
	//read files from this directory
	files, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	//shows client info for the files they want to view
	filedata := "File Name | File Size\n---------------------\n"

	/*
	* loop through each file in the directory
	* and append their names to the filenames variable
	 */

	for _, file := range files {
		//only get data for files, not directories
		if !file.Type().IsDir() {
			filename := file.Name()
			filesize := GetFileSize(filename)
			filedata += fmt.Sprintln(filename, "|", filesize)
		}
	}
	return filedata
}

func HandleConnection(c net.Conn) {
	fmt.Println(c.LocalAddr().String(), "successfully connected to server!")
	//buffer to store data coming from client
	buffer := make([]byte, 255)

	//read data sent from client
	length, err := c.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}
	/*
	* buffer is a 255 byte long array
	* that will receive data from a client
	* and length is the length of the received
	* byte array
	*
	* command is the byte array from the client
	* received but only up to length bytes since
	* we want just the data of the exact size
	 */
	command := string(buffer[:length])

	//simply show what the client picked
	fmt.Printf("Client picked the %s command", command)
	if command == "vf" {
		//send names of files in this directory to client
		c.Write([]byte(GetFileData()))
	}
}

func main() {
	//configure and start tcp server
	port := 8080
	fmt.Printf("Server running on port: %d\n", port)

	//run server on localhost
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}

	//accept client connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		//goroutine to handle incoming client connections
		go HandleConnection(conn)
	}
}
