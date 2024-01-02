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
			fmt.Println("Invalid command.\n")
		} else {
			break
		}
	}
	//send client option to server only if it's a valid one
	conn.Write([]byte(option))
}
