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
	IP_Address        string `json:"ip_address"`
	Port              string `json:"port"`
	MaxFileBufferSize int64  `json:"max_file_buffer_size"`
}

// max file buffer size to receive or upload
var maxFileBufferSize int64

// used to prepare a file to send to server
func GetFileReader(fileName string) (*os.File, *bufio.Reader, error) {
	//open file (doesn't take up memory)
	file, err := os.Open(fileName)

	//set opened file's data to reader buffer
	reader := bufio.NewReader(file)
	return file, reader, err
}

// used to make a new file for receiving (downloading)
func MakeNewFile(fileName string) *os.File {
	newFile, err := os.Create(fileName)
	if err != nil {
		fmt.Println("File make error:", err)
	}
	return newFile
}

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

	//read name of file to download
	var fileName string
	fileNameReader := bufio.NewReader(os.Stdin)
	fmt.Print("Pick a file to download: ")
	fileName, _ = fileNameReader.ReadString('\n')

	//remove two newline characters
	fileName = string(fileName[:len(fileName)-2])

	//send file name to server to download
	conn.Write([]byte(fileName))

	//make new file for file downloading
	newFile := MakeNewFile(fileName)

	//close file at the end
	defer newFile.Close()

	//buffer for server msg to see if they can send a file
	serverMessageBuffer := make([]byte, 255)

	//see if the client found a file
	length, err = conn.Read(serverMessageBuffer)
	if err != nil {
		fmt.Println(err)
	}

	//message from client to see if they can send a file
	clientMessage := string(serverMessageBuffer[:length])

	if clientMessage == "Can't read given file." {
		fmt.Printf("Couldn't open the file: %s", fileName)
	} else {
		//receive file upload from client
		DownloadFileChunks(conn, newFile, maxFileBufferSize)
		fmt.Println("Received file upload from client.")
	}

	/*

		//make new file for file downloading
		newFile := MakeNewFile(fileName)

		//receive file upload from client
		DownloadFileChunks(conn, newFile, maxFileBufferSize)

		fmt.Println("Download complete.")
	*/
}

// memory efficient file downloading using chunks of file data
// used when client wants to download a file
func DownloadFileChunks(conn net.Conn, newFile *os.File, fileBufferSize int64) {
	fileTransferIncomplete := true

	for fileTransferIncomplete {
		//empty buffer of a given size
		fileBuffer := make([]byte, fileBufferSize)

		//receive file chunk data
		length, err := conn.Read(fileBuffer)
		if err != nil {
			fmt.Println(err)
		}

		//show client progress of download
		fmt.Printf("Downloaded %d bytes\n", length)

		/*
		* check if the last 22 bytes (converted to characters)
		* make up the a message from the server to notify
		* the last n bytes were sent to stop downloading
		* and end the for loop
		 */
		if string(fileBuffer[:length][length-22:]) == "Finished sending file." {
			fileTransferIncomplete = true

			/*
			* write final received data to file
			* except the last 22 bytes (message from server)
			*
			 */
			_, err = newFile.Write(fileBuffer[:length-22])
			if err != nil {
				fmt.Println("Write error:", err)
			}
			break
		} else {
			//write the next received n bytes to file
			_, err = newFile.Write(fileBuffer[:length])
			if err != nil {
				fmt.Println("Write error:", err)
			}
		}
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

	//read file name
	var fileName string
	fileNameReader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter name of file in this folder to upload: ")
	fileName, _ = fileNameReader.ReadString('\n')

	//remove two newline characters
	fileName = string(fileName[:len(fileName)-2])

	//send name of file to server
	conn.Write([]byte(fileName))

	//get bufio reader to read file in chunks
	file, reader, err := GetFileReader(fileName)

	//if file doesn't exist on clientside, notify them
	if err != nil {
		fmt.Printf("Couldn't read the file: %s\n", fileName)
		conn.Write([]byte("Can't read given file."))
	} else {
		//tell server the client found a file
		conn.Write([]byte("This file exists."))

		//close file at the end
		defer file.Close()

		//otherwise send file to server if it exists
		SendFileChunks(conn, reader, maxFileBufferSize)
		fmt.Println("Finished uploading file.")
	}
}

// memory efficient file transferring using chunks of file data
func SendFileChunks(conn net.Conn, reader *bufio.Reader, fileBufferSize int64) {
	for {
		//empty buffer of a given size
		fileBuffer := make([]byte, fileBufferSize)
		//read every x bytes (based on fileBufferSize) into buffer until EOF
		n, err := reader.Read(fileBuffer)

		//check for errors when reading buffer
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println("Read error:", err)
		}

		//show client progress of file upload
		fmt.Printf("Uploaded %d bytes\n", n)

		//check for EOF (end of file)
		if err == nil && int64(n) < fileBufferSize {
			/*
			* send the final n bytes and a simple
			* message to notify the client
			* that all the bytes were sent
			 */
			finalBytes := string(fileBuffer[:n])
			msg := "Finished sending file."
			finalData := []byte(finalBytes + msg)

			//send the final n bytes
			_, err := conn.Write(finalData)
			if err != nil {
				fmt.Println("Write error:", err)
			}
			break
		} else {
			//send n bytes to client
			_, err := conn.Write(fileBuffer[:n])
			if err != nil {
				fmt.Println("Write error:", err)
			}
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
	maxFileBufferSize = server.MaxFileBufferSize

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
		fmt.Println("Leaving server in 2 seconds...")
		time.Sleep(time.Second * 2)
		conn.Close()
	}
}
