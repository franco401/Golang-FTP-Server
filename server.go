package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	s "strings"
	"units"
)

// returns a message showing the size of a file
func GetFileSize(fileName string) string {
	//read file
	file, err := os.Open("./files/" + fileName)
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
	files, err := os.ReadDir("./files/")
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
func SendFileData(c net.Conn) {
	/*
	* send names of files and their sizes in this directory
	* to client
	 */
	c.Write([]byte(PrepareFileData()))

	//buffer to store file name from client
	buffer := make([]byte, 255)

	//read file name sent from client they want to download
	length, err := c.Read(buffer)
	if err != nil {
		fmt.Println(err)
	}
	fileName := string(buffer[:length])

	//try to open file
	data, err := os.ReadFile("./files/" + fileName)
	if err != nil {
		/*
		* when a given file can't be read or
		* doesn't exist on the server, send
		* the client the string "error"
		 */
		c.Write([]byte("error"))
	} else {
		//send file data to client
		fmt.Println("Sending file:", fileName)
		_, err = c.Write(data)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("File successfully sent.")
		}
	}
}

// receives a file uploaded by a client
func ReceiveFileData(c net.Conn) {
	/*
	* receive file name from client (255 character)
	* after they type in the "uf" command
	 */
	fileNameLimit := 255
	fileNameBuffer := make([]byte, fileNameLimit)
	length, err := c.Read(fileNameBuffer)
	if err != nil {
		fmt.Println(err)
	}

	fileName := string(fileNameBuffer[:length])

	//receive file data from client (50 MB limit)
	fileBufferLimit := units.MB * 50
	fileBuffer := make([]byte, fileBufferLimit)
	length, err = c.Read(fileBuffer)
	if err != nil {
		fmt.Println(err)
	} else {
		//upload file
		err = os.WriteFile("./files/"+fileName, fileBuffer[:length], 0644)
		if err != nil {
			fmt.Println(err)
			//send client error message
			c.Write([]byte("File upload error."))
		} else {
			//send client sucesss message
			c.Write([]byte("File uploaded successfully."))
		}
	}
}

func HandleConnection(c net.Conn) {
	fmt.Println(c.LocalAddr().String(), "successfully connected to server!")
	//buffer to store command coming from client
	buffer := make([]byte, 10)

	//read command sent from client
	length, err := c.Read(buffer)
	if err != nil {
		fmt.Println(err)
	}
	/*
	* buffer is a byte array
	* that will receive data from a client
	* and length is the length of the received
	* byte array
	*
	* command is the byte array from the client
	* received but only up to length bytes since
	* we want just the data of the exact size
	 */
	command := string(buffer[:length])

	switch command {
	//when client picks view files command
	case "vf":
		/*
		* handles both sending user names and sizes
		* of all files on server and then sending
		* a given file to the client they want to download
		 */
		SendFileData(c)

	//when client picks upload file command
	case "uf":
		ReceiveFileData(c)
	}
}

var fileSizeLimit int = 0

func main() {
	//configure and start tcp server
	port := 8080
	fmt.Printf("Server running on port: %d\n\n", port)

	//run server on localhost
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println(err)
	}

	//read input for file size
	fileSizeReader := bufio.NewReader(os.Stdin)
	fmt.Printf("File sizes:\n---------------\nKB = kilobytes\nMB = megabyes\nGB = gigabytes\n---------------\nEnter file size: ")
	fileSize, _ := fileSizeReader.ReadString('\n')

	//remove two newline characters
	fileSize = string(fileSize[:len(fileSize)-2])

	//capitalize file size input
	fileSize = s.ToUpper(fileSize)

	//read input for amount
	amountReader := bufio.NewReader(os.Stdin)
	fmt.Printf("How many %s do you want to limit to?: ", fileSize)
	amountInput, _ := amountReader.ReadString('\n')

	//remove two newline characters
	amountInput = string(amountInput[:len(amountInput)-2])

	//convert to int
	amount, err := strconv.Atoi(amountInput)

	switch fileSize {
	case "KB":
		fileSizeLimit = units.KB * amount
	case "MB":
		fileSizeLimit = units.MB * amount
	case "GB":
		fileSizeLimit = units.GB * amount
	}

	fmt.Printf("\nFile size limit: %d %s (%d bytes)\n", amount, fileSize, fileSizeLimit)

	//accept client connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}
		//goroutine to handle incoming client connections
		go HandleConnection(conn)
	}
}
