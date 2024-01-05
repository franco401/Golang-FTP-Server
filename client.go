package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

/*
* client side functionality while communicating with
* the server
 */
func ConnectToServer(conn net.Conn) {
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
		//send client command to server only if it's a valid one
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

		//receive file data from server (50 MB limit)
		file_buffer_limit := 1048576 * 50
		file_buffer := make([]byte, file_buffer_limit)
		length, err = conn.Read(file_buffer)
		if err != nil {
			log.Fatal(err)
		}
		/*
		* if the server returns an error when reading
		* a given file, show it to the client
		 */

		if string(file_buffer[:length]) == "error" {
			fmt.Printf("Couldn't find file: %s\n", filename)
		} else {
			//else download file
			err = os.WriteFile(filename, file_buffer[:length], 0644)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Downloaded file successfully.")
		}

	} else {
		//close connection if client chooses to exit
		fmt.Println("Successfully left the server.")
		conn.Close()
	}
}

func main() {
	//user can enter ip address and port of server
	var address, port string

	fmt.Println("Enter ip address:")
	fmt.Scan(&address)

	fmt.Println("Enter port:")
	fmt.Scan(&port)

	//put the ip and port together
	server_address := address + ":" + port

	//attempt to connect to tcp server
	conn, err := net.Dial("tcp", server_address)
	if err != nil {
		fmt.Println("Couldn't connect to address:", server_address)
	} else {
		//only connect if ip and port are valid
		ConnectToServer(conn)
	}
}
