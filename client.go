package main

import (
	"fmt"
	"log"
	"net"
	"os"
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
	commands := make(map[string]int8)
	commands["vf"] = 1
	commands["e"] = 1

	/*
	* continuously loop the program until the client
	* selects a valid command
	 */
	var command string
	for {
		fmt.Println("Select option:\nvf = view files\ne = exit\n")
		//read command from user
		fmt.Scan(&command)

		if commands[command] != 1 {
			fmt.Printf("'%s' is not a valid command.\n", command)
		} else {
			break
		}
	}

	if command != "e" {
		//send client option to server only if it's a valid one
		conn.Write([]byte(command))

		//receive file names from server (10 KB limit)
		buffer := make([]byte, 10240)
		length, err := conn.Read(buffer)
		if err != nil {
			log.Fatal(err)
		}

		//read files from server
		files := string(buffer[:length])
		fmt.Println("Files on server:\n\n" + files)

		//client enters file they wish to download
		var filename string
		fmt.Println("Pick a file to download:")
		fmt.Scan(&filename)

		//send filename to server to download
		conn.Write([]byte(filename))

		//receive file data from server (700 MB limit)
		file_buffer_limit := 1048576 * 700
		file_buffer := make([]byte, file_buffer_limit)
		length, err = conn.Read(file_buffer)
		if err != nil {
			log.Fatal(err)
		}
		if length < file_buffer_limit {
			//download file
			err = os.WriteFile(filename, file_buffer[:length], 0644)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Downloaded file successfully.")
		} else {
			fmt.Println("File is too large to download.")
		}

	} else {
		//close connection if client chooses to exit
		fmt.Println("Successfully left the server.")
		conn.Close()
	}
}
