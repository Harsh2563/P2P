package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	serverAddr := "localhost:3000" // Replace with the server's IP address

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter file path (or 'exit' to quit): ")
		filePath, _ := reader.ReadString('\n')
		filePath = strings.TrimSpace(filePath)

		if filePath == "exit" {
			break
		}

		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("Failed to open file:", err)
			continue
		}

		fileInfo, err := file.Stat()
		if err != nil {
			fmt.Println("Failed to get file info:", err)
			file.Close()
			continue
		}

		fileName := fileInfo.Name()
		fileSize := fileInfo.Size()

		// Send the file name size
		err = writeInt64(conn, int64(len(fileName)))
		if err != nil {
			fmt.Println("Failed to send file name size:", err)
			file.Close()
			continue
		}

		// Send the file name
		if err != nil {
			fmt.Println("Failed to send file name:", err)
			file.Close()
			continue
		}

		// Send the file size
		err = writeInt64(conn, fileSize)
		if err != nil {
			fmt.Println("Failed to send file size:", err)
			file.Close()
			continue
		}

		// Send the file data
		_, err = io.Copy(conn, file)
		if err != nil {
			fmt.Println("Failed to send file data:", err)
			file.Close()
			continue
		}

		file.Close()
		fmt.Println("File sent successfully")
	}
}

func writeInt64(conn net.Conn, n int64) error {
	buf := []byte{
		byte(n),
		byte(n >> 8),
		byte(n >> 16),
		byte(n >> 24),
		byte(n >> 32),
		byte(n >> 40),
		byte(n >> 48),
		byte(n >> 56),
	}
	_, err := conn.Write(buf)
	return err
}
