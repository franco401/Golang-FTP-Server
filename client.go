package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"time"
)

// create json struct
type serverConfig struct {
	IP_Address        string `json:"ip_address"`
	Port              string `json:"port"`
	MaxFileBufferSize uint64 `json:"max_file_buffer_size"`
}

// max file buffer size to receive or upload
var maxFileBufferSize uint64

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

	//receive file names and sizes from server
	fileNamesAndSizesBuffer := make([]byte, 1048576)
	length, err := conn.Read(fileNamesAndSizesBuffer)
	if err != nil {
		log.Fatal(err)
	}

	//get file names from server
	fileNamesAndSizes := string(fileNamesAndSizesBuffer[:length])
	fmt.Println("Files on server:\n\n" + fileNamesAndSizes)

	//read name of file to download
	var fileName string
	fileNameReader := bufio.NewReader(os.Stdin)
	fmt.Print("Pick a file to download: ")
	fileName, _ = fileNameReader.ReadString('\n')

	if runtime.GOOS == "windows" {
		//remove two newline characters for windows
		fileName = string(fileName[:len(fileName)-2])
	} else {
		//remove one newline character otherwise
		fileName = string(fileName[:len(fileName)-1])
	}

	//send file name to server to download
	conn.Write([]byte(fileName))

	//buffer for server msg to see if they can send a file
	serverMessageBuffer := make([]byte, 255)

	//see if the server found the file
	length, err = conn.Read(serverMessageBuffer)
	if err != nil {
		fmt.Println(err)
	}

	//message from server to see if they can send the given file
	serverMessage := string(serverMessageBuffer[:length])

	//check if the server was able to read the given file
	if serverMessage == "Can't read given file." {
		fmt.Printf("Server couldn't read or find the file: %s\n", fileName)
	} else {
		//make new file for file downloading
		newFile := MakeNewFile(fileName)

		//close file at the end
		defer newFile.Close()

		//download file from server
		DownloadFileChunks(conn, newFile)
		fmt.Println("Received file upload from client.")
	}
}

// memory efficient file downloading using chunks of file data
// used when client wants to download a file
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
			* show client the last bytes downloaded
			* length-22 is used because the server
			* sends 22 extra bytes to tell the client
			* the file download is complete
			 */
			fmt.Printf("Downloaded %d bytes\n", length-22)

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
			//show client current amount of bytes downloaded
			fmt.Printf("Downloaded %d bytes\n", length)

			//write the next received n bytes to file
			_, err = newFile.Write(fileBuffer[:length])
			if err != nil {
				fmt.Println("Write error:", err)
			}
		}
	}
}

func ShowFileSize(fileSize uint64) string {
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

// returns a message showing the size of a file
func GetFileSize(fileName string) string {
	//read file
	file, err := os.Open(fileName)

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

// returns the name and size of all files on same directory as client
func PrepareFileData() string {
	//read files from same directory as client
	files, err := os.ReadDir(".")
	if err != nil {
		fmt.Println(err)
	}

	fileNamesAndSizes := ""

	for i := 0; i < len(files); i++ {
		//only get data for files, not directories
		if !files[i].Type().IsDir() {
			//string formatted as: [filename] | [filesize]
			fileNamesAndSizes += fmt.Sprint(files[i].Name(), " | ", GetFileSize(files[i].Name()))
		}
	}
	return fileNamesAndSizes
}

// called when client picks upload file command
func UploadFile(command string, conn net.Conn) {
	//send command to server
	conn.Write([]byte(command))

	//show list of files in directory first
	clientSideFiles := PrepareFileData()
	fmt.Println("Pick a file to upload:\n\n" + clientSideFiles)

	//read file name
	var fileName string
	fileNameReader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter name of file in this folder to upload: ")
	fileName, _ = fileNameReader.ReadString('\n')

	if runtime.GOOS == "windows" {
		//remove two newline characters for windows
		fileName = string(fileName[:len(fileName)-2])
	} else {
		//remove one newline character otherwise
		fileName = string(fileName[:len(fileName)-1])
	}

	//send name of file to server to download
	conn.Write([]byte(fileName))

	//get bufio reader to read file in chunks
	file, reader, err := GetFileReader(fileName)

	//if file doesn't exist on clientside, notify them
	if err != nil {
		fmt.Printf("Couldn't read the file: %s\n", fileName)
		conn.Write([]byte("Can't read given file."))
	} else {
		//close file at the end
		defer file.Close()

		//otherwise send file to server if it exists
		SendFileChunks(conn, reader)
		fmt.Println("Finished uploading file.")
	}
}

// memory efficient file transferring using chunks of file data
func SendFileChunks(conn net.Conn, reader *bufio.Reader) {
	for {
		//empty buffer of a given size
		fileBuffer := make([]byte, maxFileBufferSize)
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

		if runtime.GOOS == "windows" {
			//remove two newline characters for windows
			command = string(command[:len(command)-2])
		} else {
			//remove one newline character otherwise
			command = string(command[:len(command)-1])
		}

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
		fmt.Println("Closing client...")
		time.Sleep(time.Second)
	} else {
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
		} else {
			//only connect if ip and port are valid
			ConnectToServer(conn)
			fmt.Println("Leaving server in 3 seconds...")
			time.Sleep(time.Second * 3)
			conn.Close()
		}
	}
}
