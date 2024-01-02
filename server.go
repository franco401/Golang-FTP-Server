package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

func GetFileNames() string {
	//read files from this directory
	files, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	var filenames string
	/*
	* loop through each file in the directory
	* and append their names to the filenames variable
	 */
	for _, file := range files {
		filename := file.Name()
		filenames += filename + "\n"
	}
	return filenames
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
		c.Write([]byte(GetFileNames()))
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
