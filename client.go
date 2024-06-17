package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	serverAddr := "192.168.1.100:3000" // Replace with the server's IP address
	filePath := "path/to/your/file"    // Update this path to the file you want to send

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		return
	}
	defer conn.Close()

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Failed to open file:", err)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Failed to get file info:", err)
		return
	}

	fileName := fileInfo.Name()
	fileSize := fileInfo.Size()

	// Send the file name size
	err = writeInt64(conn, int64(len(fileName)))
	if err != nil {
		fmt.Println("Failed to send file name size:", err)
		return
	}

	// Send the file name
	_, err = conn.Write([]byte(fileName))
	if err != nil {
		fmt.Println("Failed to send file name:", err)
		return
	}

	// Send the file size
	err = writeInt64(conn, fileSize)
	if err != nil {
		fmt.Println("Failed to send file size:", err)
		return
	}

	// Send the file data
	_, err = io.Copy(conn, file)
	if err != nil {
		fmt.Println("Failed to send file data:", err)
		return
	}

	fmt.Println("File sent successfully")
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
