package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

// create json struct
type serverConfig struct {
	IP_Address           string `json:"ip_address"`
	Port                 string `json:"port"`
	MaxFileBufferSize    uint64 `json:"max_file_buffer_size"`
	FileStorageDirectory string `json:"file_storage_directory"`
}

// max file buffer size to receive or upload
var maxFileBufferSize uint64

// location of where files are stored and read from for server
var fileStorageDirectory string

// used to prepare a file to send to client
func GetFileReader(fileName string) (*os.File, *bufio.Reader, error) {
	//open file (doesn't take up memory)
	file, err := os.Open(fileStorageDirectory + fileName)
	if err != nil {
		fmt.Println("File read error:", err)
	}

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

// returns a message showing the size of a file
func GetFileSize(fileName string) string {
	//read file
	file, err := os.Open(fileStorageDirectory + fileName)

	//close file at the end of current file goroutine
	defer file.Close()

	if err != nil {
		fmt.Println(err)
	}

	f, err := file.Stat()

	if err != nil {
		fmt.Println(err)
	}

	//file size in bytes
	fs := f.Size()
	var fileSize float64 = float64(fs)

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

// returns the name and size of all files on server to client
func PrepareFileData() string {
	//read files from this directory
	files, err := os.ReadDir(fileStorageDirectory)
	if err != nil {
		fmt.Println(err)
	}

	//shows client info for the files on the server
	fileData := "File Name | File Size\n---------------------\n"

	/*
	* loop through each file in the directory
	* and append their names to the file names variable
	 */

	for _, file := range files {
		//only get data for files, not directories
		if !file.Type().IsDir() {
			fileName := file.Name()
			fileSize := GetFileSize(fileName)
			fileData += fmt.Sprintln(fileName, "|", fileSize)
		}
	}
	return fileData
}

// reads a given file name and sends it to client
func SendFileData(conn net.Conn) {
	/*
	* send names of files and their sizes in this directory
	* to client
	 */
	conn.Write([]byte(PrepareFileData()))

	//buffer to store file name from client
	fileNameBuffer := make([]byte, 255)

	//read file name sent from client they want to download
	length, err := conn.Read(fileNameBuffer)
	if err != nil {
		fmt.Println(err)
	}
	fileName := string(fileNameBuffer[:length])

	//get bufio reader to read file in chunks
	file, reader, err := GetFileReader(fileName)

	//if file doesn't exist on serverise, notify client
	if err != nil {
		conn.Write([]byte("Can't read given file."))
	} else {
		//tell client the server found a file
		conn.Write([]byte("This file exists."))

		//close file at the end
		defer file.Close()

		//otherwise send file to server if it exists
		SendFileChunks(conn, reader)
		fmt.Println("Finished sending file to client.")
	}
}

// memory efficient file transferring using chunks of file data
// used when client wants to download a file
func SendFileChunks(conn net.Conn, reader *bufio.Reader) {
	for {
		//empty buffer of a given size
		fileBuffer := make([]byte, maxFileBufferSize)
		//read every x bytes (based on maxFileBufferSize) into buffer until EOF
		n, err := reader.Read(fileBuffer)

		//check for errors when reading buffer
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println("Read error:", err)
		}

		//check for EOF (end of file)
		if err == nil && uint64(n) < maxFileBufferSize {
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

// receives a file uploaded by a client
func ReceiveFileData(conn net.Conn) {
	/*
	* receive file name from client (255 character)
	* after they type in the "uf" command
	 */
	fileNameLimit := 255
	fileNameBuffer := make([]byte, fileNameLimit)
	length, err := conn.Read(fileNameBuffer)
	if err != nil {
		fmt.Println(err)
	}

	//read file name client wants to upload
	fileName := string(fileNameBuffer[:length])

	//buffer for client msg to see if they can send a file
	clientMessageBuffer := make([]byte, 255)

	//see if the client found a file
	length, err = conn.Read(clientMessageBuffer)
	if err != nil {
		fmt.Println(err)
	}

	//message from client to see if they can send a file
	clientMessage := string(clientMessageBuffer[:length])

	//check if the client was able to read a file to upload here
	if !(clientMessage == "Can't read given file.") {
		//prepare new file (to receive file upload)
		newFile := MakeNewFile(fileStorageDirectory + fileName)

		//close file at the end
		defer newFile.Close()

		//receive file upload from client
		DownloadFileChunks(conn, newFile)
		fmt.Println("Received file upload from client.")
	}
}

// memory efficient file downloading using chunks of file data
// used when client wants to upload a file to here
func DownloadFileChunks(conn net.Conn, newFile *os.File) {
	fileTransferIncomplete := true

	for fileTransferIncomplete {
		//empty buffer of a given size
		fileBuffer := make([]byte, maxFileBufferSize)

		//receive file chunk data
		length, err := conn.Read(fileBuffer)
		if err != nil {
			fmt.Println(err)
		}

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

func HandleConnection(conn net.Conn) {
	fmt.Println(conn.LocalAddr().String(), "successfully connected to server!")

	//buffer to store command coming from client
	commandBuffer := make([]byte, 10)

	//read command sent from client
	length, err := conn.Read(commandBuffer)
	if err != nil {
		fmt.Println(err)
	}
	/*
	* commandBuffer is a byte array
	* that will receive data from a client
	* and length is the length of the received
	* byte array
	*
	* command is the byte array from the client
	* received but only up to length bytes since
	* we want just the data of the exact size
	 */
	command := string(commandBuffer[:length])

	switch command {
	//when client picks view files command
	case "vf":
		/*
		* handles both sending user names and sizes
		* of all files on server and then sending
		* a given file to the client they want to download
		 */
		SendFileData(conn)

	//when client picks upload file command
	case "uf":
		ReceiveFileData(conn)
	}
}

func main() {
	//read json file
	data, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Closing server...")
		time.Sleep(time.Second)
	}

	//use json struct
	var server serverConfig

	//and give it the values of the json file
	err = json.Unmarshal(data, &server)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Closing server...")
		time.Sleep(time.Second)
	}

	//set file buffer limit
	maxFileBufferSize = server.MaxFileBufferSize

	if maxFileBufferSize < 1024 {
		fmt.Printf("max_file_buffer_size of %d bytes is too small.\n", maxFileBufferSize)
		fmt.Println("Closing server...")
		time.Sleep(time.Second)
	} else {
		//set server file storage directory
		fileStorageDirectory = server.FileStorageDirectory

		//read ip address and port
		serverAddress := server.IP_Address + ":" + server.Port

		//run server from server config and start tcp server
		listener, err := net.Listen("tcp", serverAddress)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Closing server...")
			time.Sleep(time.Second)
		} else {
			fmt.Printf("Server running on: %s\n", serverAddress)

			//accept client connections
			for {
				conn, err := listener.Accept()
				if err != nil {
					fmt.Println(err)
					fmt.Println("Closing server...")
					time.Sleep(time.Second)
				}
				//goroutine to handle incoming client connections
				go HandleConnection(conn)
			}
		}
	}
}
