package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

// when client picks view files command
func ViewFiles(command string, conn net.Conn) {
	conn.Write([]byte(command))

	//receive file names from server (10 KB limit)
	buffer := make([]byte, 10240)
	length, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	//get file names from server
	files := string(buffer[:length])
	fmt.Println("Files on server:\n\n" + files)

	//user input for filename to download
	var filename string
	filename_reader := bufio.NewReader(os.Stdin)
	fmt.Print("Pick a file to download: ")
	filename, _ = filename_reader.ReadString('\n')

	//remove two newline characters
	filename = string(filename[:len(filename)-2])

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
		//since the program ends here for now, close connection
		fmt.Println("Successfully left the server.")
		conn.Close()
	}
}

// will be updated in later
func UploadFile(command string, conn net.Conn) {
	fmt.Println(command)
}

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
	commands["uf"] = 1
	commands["e"] = 1

	/*
	* continuously loop the program until the client
	* selects a valid command
	 */
	var command string
	for {
		fmt.Println("Commands:\n---------\nvf = view files\nuf = upload file\ne = exit\n")

		/*
			fmt.Print("Select command: ")
			//read command from user
			fmt.Scan(&command)
		*/

		command_reader := bufio.NewReader(os.Stdin)
		fmt.Print("Select command: ")
		command, _ = command_reader.ReadString('\n')

		//remove two newline characters
		command = string(command[:len(command)-2])

		if commands[command] != 1 {
			fmt.Printf("'%s' is not a valid command.\n\n", command)
		} else {
			break
		}
	}

	switch command {
	case "vf":
		//send vf command to server
		ViewFiles(command, conn)

	case "uf":
		//send uf command to server
		UploadFile(command, conn)

	//when client picks exit command
	case "e":
		//close connection if client chooses to exit
		fmt.Println("Successfully left the server.")
		conn.Close()
	}

}

func main() {
	//user can enter ip address and port of server
	var address, port string

	address_reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter ip address: ")
	address, _ = address_reader.ReadString('\n')

	//remove two newline characters
	address = string(address[:len(address)-2])

	port_reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter port: ")
	port, _ = port_reader.ReadString('\n')

	//remove two newline characters
	port = string(port[:len(port)-2])

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
