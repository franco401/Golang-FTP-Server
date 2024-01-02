package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

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
	* receivedData is the byte array from the client
	* received but only up to length bytes since
	* we want just the data of the exact size
	 */
	receivedData := string(buffer[:length])
	fmt.Println("Client picked: ", receivedData)
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
