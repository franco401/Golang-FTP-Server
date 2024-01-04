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

	//command options a client can pick
	options := make(map[string]int8)
	options["vf"] = 1
	options["e"] = 1

	/*
	* continuously loop the program until the client
	* selects a valid command
	 */
	var option string
	for {
		fmt.Println("Select option:\nvf = view files\ne = exit\n")
		//read option from user
		fmt.Scan(&option)

		if options[option] != 1 {
			fmt.Printf("'%s' is not a valid command.\n", option)
		} else {
			break
		}
	}

	if option != "e" {
		//send client option to server only if it's a valid one
		conn.Write([]byte(option))

		//receive file names from server directory
		buffer := make([]byte, 1024)
		length, err := conn.Read(buffer)
		if err != nil {
			log.Fatal(err)
		}
		//read file names from server
		files := string(buffer[:length])
		fmt.Println("Files on server:\n", files)
	} else {
		//close connection if client chooses to exit
		fmt.Println("Successfully left the server.")
		conn.Close()
	}
}
