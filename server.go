package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"units"
)

// returns a message showing the size of a file
func GetFileSize(filename string) string {
	//read file
	file, err := os.Open("./files/" + filename)
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
	var filesize float64 = float64(fs)

	filesizes := []string{"B", "KB", "MB", "GB"}
	var filesize_index int8

	/*
	* while the filesize in bytes
	* is bigger than 1024, continously
	* divide by 1024 and increment the
	* filesize_index to finally return
	* the appropriate file size for the
	* client, such as 1 KB for 1024 bytes
	 */
	for filesize >= 1024.0 {
		filesize_index++
		filesize /= 1024.0
	}
	return fmt.Sprintln(filesize, filesizes[filesize_index])
}

// returns the name and size of all files on server to client
func PrepareFileData() string {
	//read files from this directory
	files, err := os.ReadDir("./files/")
	if err != nil {
		fmt.Println(err)
	}

	//shows client info for the files on the server
	filedata := "File Name | File Size\n---------------------\n"

	/*
	* loop through each file in the directory
	* and append their names to the filenames variable
	 */

	for _, file := range files {
		//only get data for files, not directories
		if !file.Type().IsDir() {
			filename := file.Name()
			filesize := GetFileSize(filename)
			filedata += fmt.Sprintln(filename, "|", filesize)
		}
	}
	return filedata
}

// reads a given filename and sends it to client
func SendFileData(c net.Conn) {
	/*
	* send names of files and their sizes in this directory
	* to client
	 */
	c.Write([]byte(PrepareFileData()))

	//buffer to store file name from client
	buffer := make([]byte, 255)

	//read filename sent from client they want to download
	length, err := c.Read(buffer)
	if err != nil {
		fmt.Println(err)
	}
	filename := string(buffer[:length])

	//try to open file
	data, err := os.ReadFile("./files/" + filename)
	if err != nil {
		/*
		* when a given file can't be read or
		* doesn't exist on the server, send
		* the client the string "error"
		 */
		c.Write([]byte("error"))
	} else {
		//send file data to client
		fmt.Println("Sending file:", filename)
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
	* receive filename from client (255 character)
	* after they type in the "uf" command
	 */
	filename_limit := 255
	filename_buffer := make([]byte, filename_limit)
	length, err := c.Read(filename_buffer)
	if err != nil {
		fmt.Println(err)
	}

	filename := string(filename_buffer[:length])

	//receive file data from client (50 MB limit)
	file_buffer_limit := units.MB * 50
	file_buffer := make([]byte, file_buffer_limit)
	length, err = c.Read(file_buffer)
	if err != nil {
		fmt.Println(err)
	} else {
		//upload file
		err = os.WriteFile("./files/"+filename, file_buffer[:length], 0644)
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

func main() {
	//configure and start tcp server
	port := 8080
	fmt.Printf("Server running on port: %d\n", port)

	//run server on localhost
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println(err)
	}

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
