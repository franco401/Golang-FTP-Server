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
	fmt.Println("Successfully connected to server.\n")

	//give client options to pick
	var option string
	fmt.Println("Select option:\n vf = view files\ne = exit\n")
	//read option from user
	fmt.Scan(&option)

	//send client option to server
	conn.Write([]byte(option))
}
