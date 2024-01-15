package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"units"
)

// create json struct
type serverConfig struct {
	IP_Address      string `json:"ip_address"`
	Port            string `json:"port"`
	FileBufferLimit int64  `json:"file_buffer_limit"`
}

var fileBufferLimit int64

// called when client picks view files command
func ViewFiles(command string, conn net.Conn) {
	//send command to server
	conn.Write([]byte(command))

	//receive file names from server (10 KB limit)
	fileNamesBuffer := make([]byte, units.KB*10)
	length, err := conn.Read(fileNamesBuffer)
	if err != nil {
		log.Fatal(err)
	}

	//get file names from server
	files := string(fileNamesBuffer[:length])
	fmt.Println("Files on server:\n\n" + files)

	//show client the file download limit from config.json (file buffer limit)
	fmt.Printf("File download limit: %s\n", ShowFileSize(fileBufferLimit))

	//read name of file to download
	var fileName string
	fileNameReader := bufio.NewReader(os.Stdin)
	fmt.Print("Pick a file to download: ")
	fileName, _ = fileNameReader.ReadString('\n')

	//remove two newline characters
	fileName = string(fileName[:len(fileName)-2])

	//send file name to server to download
	conn.Write([]byte(fileName))

	//file download limit (50 MB limit)
	//fileBufferLimit := units.MB * 50

	//receive file data from server to download
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
		fmt.Println("Exiting server...")
		time.Sleep(time.Second)
		conn.Close()
	} else {
		//else download file
		err = os.WriteFile(fileName, fileBuffer[:length], 0644)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Downloaded file successfully.")
		fmt.Println("Exiting server...")
		time.Sleep(time.Second)
		conn.Close()
	}
}

func ShowFileSize(fileSize int64) string {
	fileSizes := []string{"B", "KB", "MB", "GB"}
	var fileSizeIndex int8

	/*
	* while the filesize in bytes
	* is bigger than 1024, continously
	* divide by 1024 and increment the
	* filesizeIndex to finally return
	* the appropriate file size for the
	* client, such as 1 KB for 1024 bytes
	 */
	for fileSize >= 1024.0 {
		fileSizeIndex++
		fileSize /= 1024.0
	}
	return fmt.Sprintln(fileSize, fileSizes[fileSizeIndex])
}

// called when client picks upload file command
func UploadFile(command string, conn net.Conn) {
	//send command to server
	conn.Write([]byte(command))

	//show client the file upload limit from config.json (file buffer limit)
	fmt.Printf("File upload limit: %s\n", ShowFileSize(fileBufferLimit))

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
		fmt.Println("Exiting server...")
		time.Sleep(time.Second)
		conn.Close()
	} else {
		//check if the chosen file is wihin the file upload limit
		if int64(len(data)) > fileBufferLimit {
			fmt.Printf("This file is too large: %s", ShowFileSize(int64(len(data))))
			fmt.Println("Exiting server...")
			time.Sleep(time.Second)
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
			fmt.Println("Exiting server...")
			time.Sleep(time.Second)
			conn.Close()
		}
	}
}

// returns a (valid) command the client picks
func CommandSelection() string {
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
	return command
}

/*
* client side functionality while communicating with
* the server
 */
func ConnectToServer(conn net.Conn) {
	//show client message after connecting
	fmt.Printf("Successfully connected to server.\n\n")

	command := CommandSelection()

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

func main() {
	//read json file
	data, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println(err)
	}

	//use json struct
	var server serverConfig

	//and give it the values of the json file
	err = json.Unmarshal(data, &server)
	if err != nil {
		fmt.Println(err)
	}

	//set file buffer limit
	fileBufferLimit = server.FileBufferLimit

	//read ip address and port
	serverAddress := server.IP_Address + ":" + server.Port

	//attempt to connect to tcp server
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Println("Couldn't connect to address:", serverAddress)
		fmt.Println("Exiting server...")
		time.Sleep(time.Second)
		conn.Close()
	} else {
		//only connect if ip and port are valid
		ConnectToServer(conn)
	}
}
