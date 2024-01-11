package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"units"
)

// called when client picks view files command
func ViewFiles(command string, conn net.Conn) {
	//send command to server
	conn.Write([]byte(command))

	//receive file names from server (10 KB limit)
	buffer := make([]byte, units.KB*10)
	length, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	//get file names from server
	files := string(buffer[:length])
	fmt.Println("Files on server:\n\n" + files)

	//read name of file to download
	var fileName string
	fileNameReader := bufio.NewReader(os.Stdin)
	fmt.Print("Pick a file to download: ")
	fileName, _ = fileNameReader.ReadString('\n')

	//remove two newline characters
	fileName = string(fileName[:len(fileName)-2])

	//send file name to server to download
	conn.Write([]byte(fileName))

	//receive file data from server (50 MB limit)
	fileBufferLimit := units.MB * 50
	fileBuffer := make([]byte, fileBufferLimit)
	length, err = conn.Read(fileBuffer)
	if err != nil {
		log.Fatal(err)
	}
	/*
	* if the server returns an error when reading
	* a given file, show it to the client
	 */

	if string(fileBuffer[:length]) == "error" {
		fmt.Printf("Couldn't find file: %s\n", fileName)
	} else {
		//else download file
		err = os.WriteFile(fileName, fileBuffer[:length], 0644)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Downloaded file successfully.")
		//since the program ends here for now, close connection
		fmt.Println("Successfully left the server.")
		conn.Close()
	}
}

// called when client picks upload file command
func UploadFile(command string, conn net.Conn) {
	//send command to server
	conn.Write([]byte(command))

	//read file name
	var fileName string
	fileNameReader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter name of file in this folder to upload: ")
	fileName, _ = fileNameReader.ReadString('\n')

	//remove two newline characters
	fileName = string(fileName[:len(fileName)-2])

	//try to open (local) file
	data, err := os.ReadFile(fileName)
	if err != nil {
		//close connection if local file can't be read
		fmt.Println(err)
		conn.Close()
	} else {
		//send name of file to server
		conn.Write([]byte(fileName))

		/*
		* give the server some time to read file data
		* after receiving file name
		 */
		time.Sleep(time.Millisecond)

		//send file data to server
		conn.Write(data)

		//receive message from server (255 characters)
		messageLimit := 255
		messageBuffer := make([]byte, messageLimit)
		length, err := conn.Read(messageBuffer)
		if err != nil {
			fmt.Println(err)
		}
		message := string(messageBuffer[:length])
		fmt.Println("Message from server:", message)
	}
}

/*
* client side functionality while communicating with
* the server
 */
func ConnectToServer(conn net.Conn) {
	//show client message after connecting
	fmt.Printf("Successfully connected to server.\n\n")

	//commands a client can pick
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
		fmt.Printf("Commands:\n---------\nvf = view files\nuf = upload file\ne = exit\n\n")
		commandReader := bufio.NewReader(os.Stdin)
		fmt.Print("Select command: ")
		command, _ = commandReader.ReadString('\n')

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
		err := conn.Close()
		if err != nil {
			fmt.Println(err)
		} else {
			//close connection if client chooses to exit
			fmt.Println("Successfully left the server.")
		}
	}

}

// user can enter ip address and port of server
func InputIPAddress() string {
	var address, port string

	addressReader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter ip address: ")
	address, _ = addressReader.ReadString('\n')

	//remove two newline characters
	address = string(address[:len(address)-2])

	portReader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter port: ")
	port, _ = portReader.ReadString('\n')

	//remove two newline characters
	port = string(port[:len(port)-2])

	//put the ip and port together
	return address + ":" + port
}

func main() {
	serverAddress := InputIPAddress()

	//attempt to connect to tcp server
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Println("Couldn't connect to address:", serverAddress)
	} else {
		//only connect if ip and port are valid
		ConnectToServer(conn)
	}
}
