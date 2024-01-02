package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

func main() {
	port := 8080

	//attempt to connect to tcp server
	conn, err := net.Dial("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}

	//show client message after connecting
	fmt.Println("Successfully connected to server.")

	data := []byte("Hello world!")
	fmt.Println("Sending data: ", string(data))

	//send data to server
	conn.Write(data)
}
